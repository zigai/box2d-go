package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var approvedReference struct {
	sync.Once

	manifest referenceManifest
	err      error
}

type referenceManifest struct {
	Schema          int                        `json:"schema"`
	Upstream        string                     `json:"upstream"`
	Precision       string                     `json:"precision"`
	Scalar          bool                       `json:"scalar"`
	Assertions      bool                       `json:"assertions"`
	Contraction     string                     `json:"contraction"`
	Compiler        string                     `json:"compiler"`
	CompilerVersion string                     `json:"compiler_version"` //nolint:tagliatelle // Fixed native C manifest key emitted by oracle/reference.cmake.
	CompilerTarget  string                     `json:"compiler_target"`  //nolint:tagliatelle // Fixed native C manifest key emitted by oracle/reference.cmake.
	BuildType       string                     `json:"build_type"`       //nolint:tagliatelle // Fixed native C manifest key emitted by oracle/reference.cmake.
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

	if err := validateReferenceProfile(root, &reference); err != nil {
		return reference, err
	}

	if err := verifyAdapterSources(root, &reference); err != nil {
		return reference, err
	}

	if len(reference.Fixtures) == 0 {
		return reference, errors.New("C provenance contains no completed fixtures") //nolint:err113,staticcheck // Local oracle test setup diagnostic; Open reports it with Fatalf. C names the native reference language; TestReferenceRejectsStalePinProfileAndAdapters covers these failures.
	}

	if err := verifyDebugReference(&reference); err != nil {
		return reference, err
	}

	return reference, nil
}

func validateReferenceProfile(root string, reference *referenceManifest) error {
	pin, err := os.ReadFile(filepath.Join(root, "UPSTREAM"))
	if err != nil {
		return fmt.Errorf("read C upstream pin: %w", err)
	}

	if reference.Schema != 1 || reference.Upstream != strings.TrimSpace(string(pin)) {
		return fmt.Errorf("stale C provenance: schema %d, upstream %q", reference.Schema, reference.Upstream) //nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
	}

	if reference.Precision != config || !reference.Scalar || !reference.Assertions ||
		reference.BuildType != "RelWithDebInfo" || reference.Contraction != "off" {
		//nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
		return fmt.Errorf("wrong C reference profile: precision %q, scalar %t, assertions %t, build type %q, contraction %q",
			reference.Precision, reference.Scalar, reference.Assertions, reference.BuildType, reference.Contraction)
	}

	return validateCompilerProvenance(reference)
}

func validateCompilerProvenance(reference *referenceManifest) error {
	if reference.Compiler == "" || reference.CompilerVersion == "" || reference.CompilerTarget == "" ||
		reference.BuildType == "" || !strings.Contains(reference.Flags, "-ffp-contract=off") {
		return errors.New("incomplete C compiler provenance") //nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
	}

	return nil
}

func verifyAdapterSources(root string, reference *referenceManifest) error {
	files, err := adapterSources(root)
	if err != nil {
		return err
	}

	if len(files) != len(reference.Sources) {
		return fmt.Errorf("stale C adapter inventory: %d source files, provenance contains %d", len(files), len(reference.Sources)) //nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
	}

	for _, name := range files {
		want, exists := reference.Sources[name]
		if !exists {
			return fmt.Errorf("C adapter %q is absent from provenance", name) //nolint:err113,staticcheck // Local oracle test setup diagnostic; Open reports it with Fatalf. C names the native reference language; TestReferenceRejectsStalePinProfileAndAdapters covers these failures.
		}

		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Errorf("read C adapter %q: %w", name, err)
		}

		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != want {
			return fmt.Errorf("stale C adapter %q", name) //nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
		}
	}

	return nil
}

func verifyDebugReference(reference *referenceManifest) error {
	if reference.Debug != nil {
		debug := reference.Debug
		if debug.Schema != reference.Schema || debug.Upstream != reference.Upstream ||
			debug.Precision != config || !debug.Scalar || debug.Contraction != "off" ||
			!debug.Assertions || debug.BuildType != "Debug" || len(debug.Sources) != len(reference.Sources) {
			return errors.New("wrong C Debug diagnostic profile") //nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
		}

		for name, digest := range reference.Sources {
			if debug.Sources[name] != digest {
				return fmt.Errorf("stale C Debug adapter %q", name) //nolint:err113 // Local oracle test setup diagnostic; Open reports it with Fatalf.
			}
		}
	}

	return nil
}

func adapterSources(root string) ([]string, error) {
	files := []string{
		"oracle/oracle.c", "oracle/oracle.h", "oracle/differential.c", "oracle/CMakeLists.txt",
		"oracle/build.sh", "oracle/reference.cmake", "oracle/capture.cmake",
		"oracle/check-upstream.cmake", "oracle/cases.json", "oracle/patches/world_overlap_stop.patch",
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
				return fmt.Errorf("resolve adapter source relative to repository: %w", err)
			}

			files = append(files, filepath.ToSlash(relative))

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk C adapter directory %q: %w", dir, err)
		}
	}

	return files, nil
}

func verifyCapture(reference *referenceManifest, name string, data []byte) (int, error) {
	identity, exists := reference.Fixtures[name]
	if !exists || !identity.Completed || identity.Producer == "" || identity.Records <= 0 {
		return 0, fmt.Errorf("fixture %q has no completed C capture", name) //nolint:err113 // Local oracle test diagnostic; Open reports it with Fatalf.
	}

	if err := verifyCaptureProfile(reference, name, identity); err != nil {
		return 0, err
	}

	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != identity.SHA256 {
		return 0, fmt.Errorf("fixture %q differs from its completed C capture", name) //nolint:err113 // Local oracle test diagnostic; Open reports it with Fatalf.
	}

	return identity.Records, nil
}

func verifyCaptureProfile(reference *referenceManifest, name string, identity fixtureIdentity) error {
	if identity.Profile == "" {
		return nil
	}

	if identity.Profile != "debug" || reference.Debug == nil {
		return fmt.Errorf("fixture %q has an unknown C diagnostic profile", name) //nolint:err113 // Local oracle test diagnostic; Open reports it with Fatalf.
	}

	debug, exists := reference.Debug.Fixtures[name]
	if !exists || !debug.Completed || debug.Records != identity.Records || debug.SHA256 != identity.SHA256 {
		return fmt.Errorf("fixture %q has inconsistent C Debug provenance", name) //nolint:err113 // Local oracle test diagnostic; Open reports it with Fatalf.
	}

	return nil
}
