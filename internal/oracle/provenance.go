package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type referenceManifest struct {
	Schema          int                        `json:"schema"`
	Upstream        string                     `json:"upstream"`
	Precision       string                     `json:"precision"`
	Scalar          bool                       `json:"scalar"`
	Assertions      bool                       `json:"assertions"`
	Contraction     string                     `json:"contraction"`
	Compiler        string                     `json:"compiler"`
	CompilerVersion string                     `json:"compiler_version"`
	CompilerTarget  string                     `json:"compiler_target"`
	BuildType       string                     `json:"build_type"`
	Flags           string                     `json:"flags"`
	Sources         map[string]string          `json:"sources"`
	Fixtures        map[string]fixtureIdentity `json:"fixtures"`
	Debug           *referenceManifest         `json:"debug,omitempty"`
}

type fixtureIdentity struct {
	Producer  string `json:"producer"`
	Records   int    `json:"records"`
	SHA256    string `json:"sha256"`
	Completed bool   `json:"completed"`
	Profile   string `json:"profile,omitempty"`
}

var approvedReference struct {
	sync.Once
	manifest referenceManifest
	err      error
}

func fixtureReference(root string) (*referenceManifest, error) {
	approvedReference.Do(func() {
		approvedReference.manifest, approvedReference.err = loadReference(root,
			filepath.Join(root, "testdata", "oracle", config, "manifest.json"))
	})
	return &approvedReference.manifest, approvedReference.err
}

func loadReference(root, file string) (referenceManifest, error) {
	var reference referenceManifest
	data, err := os.ReadFile(file)
	if err != nil {
		return reference, fmt.Errorf("cannot read C provenance: %w", err)
	}
	if err := json.Unmarshal(data, &reference); err != nil {
		return reference, fmt.Errorf("malformed C provenance: %w", err)
	}
	pin, err := os.ReadFile(filepath.Join(root, "UPSTREAM"))
	if err != nil {
		return reference, err
	}
	if reference.Schema != 1 || reference.Upstream != strings.TrimSpace(string(pin)) {
		return reference, fmt.Errorf("stale C provenance: schema %d, upstream %q", reference.Schema, reference.Upstream)
	}
	if reference.Precision != config || !reference.Scalar || !reference.Assertions ||
		reference.BuildType != "RelWithDebInfo" || reference.Contraction != "off" {
		return reference, fmt.Errorf("wrong C reference profile: precision %q, scalar %t, assertions %t, build type %q, contraction %q",
			reference.Precision, reference.Scalar, reference.Assertions, reference.BuildType, reference.Contraction)
	}
	if reference.Compiler == "" || reference.CompilerVersion == "" || reference.CompilerTarget == "" ||
		reference.BuildType == "" || !strings.Contains(reference.Flags, "-ffp-contract=off") {
		return reference, fmt.Errorf("incomplete C compiler provenance")
	}
	files, err := adapterSources(root)
	if err != nil {
		return reference, err
	}
	if len(files) != len(reference.Sources) {
		return reference, fmt.Errorf("stale C adapter inventory: %d source files, provenance contains %d", len(files), len(reference.Sources))
	}
	for _, name := range files {
		want, exists := reference.Sources[name]
		if !exists {
			return reference, fmt.Errorf("C adapter %q is absent from provenance", name)
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return reference, err
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != want {
			return reference, fmt.Errorf("stale C adapter %q", name)
		}
	}
	if len(reference.Fixtures) == 0 {
		return reference, fmt.Errorf("C provenance contains no completed fixtures")
	}
	if reference.Debug != nil {
		debug := reference.Debug
		if debug.Schema != reference.Schema || debug.Upstream != reference.Upstream ||
			debug.Precision != config || !debug.Scalar || debug.Contraction != "off" ||
			!debug.Assertions || debug.BuildType != "Debug" || len(debug.Sources) != len(reference.Sources) {
			return reference, fmt.Errorf("wrong C Debug diagnostic profile")
		}
		for name, digest := range reference.Sources {
			if debug.Sources[name] != digest {
				return reference, fmt.Errorf("stale C Debug adapter %q", name)
			}
		}
	}
	return reference, nil
}

func adapterSources(root string) ([]string, error) {
	files := []string{
		"oracle/oracle.c", "oracle/oracle.h", "oracle/differential.c", "oracle/CMakeLists.txt",
		"oracle/build.sh", "oracle/reference.cmake", "oracle/capture.cmake",
		"oracle/check-upstream.cmake", "oracle/cases.json",
	}
	for _, dir := range []string{"sample", "tests", "debug", "bench"} {
		err := filepath.WalkDir(filepath.Join(root, "oracle", dir), func(file string, entry fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) && file == filepath.Join(root, "oracle", dir) {
					return nil
				}
				return err
			}
			if entry.IsDir() || (filepath.Ext(file) != ".c" && filepath.Ext(file) != ".h") {
				return nil
			}
			relative, err := filepath.Rel(root, file)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(relative))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func verifyCapture(reference *referenceManifest, name string, data []byte) (int, error) {
	identity, exists := reference.Fixtures[name]
	if !exists || !identity.Completed || identity.Producer == "" || identity.Records <= 0 {
		return 0, fmt.Errorf("fixture %q has no completed C capture", name)
	}
	if identity.Profile != "" {
		if identity.Profile != "debug" || reference.Debug == nil {
			return 0, fmt.Errorf("fixture %q has an unknown C diagnostic profile", name)
		}
		debug, exists := reference.Debug.Fixtures[name]
		if !exists || !debug.Completed || debug.Records != identity.Records || debug.SHA256 != identity.SHA256 {
			return 0, fmt.Errorf("fixture %q has inconsistent C Debug provenance", name)
		}
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != identity.SHA256 {
		return 0, fmt.Errorf("fixture %q differs from its completed C capture", name)
	}
	return identity.Records, nil
}
