package dagjose

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/traversal"
	"github.com/warpfork/go-testmark"
)

func TestSpecFixtures(t *testing.T) {
	// Load the file.
	// Also prepare a patch accumlater in case we're in regen mode.
	file := "../.ipld/specs/codecs/dag-jose/fixtures/index.md"
	doc, err := testmark.ReadFile(file)
	if os.IsNotExist(err) {
		t.Skipf("not running spec suite: %s (did you clone the submodule with the data?)", err)
	}
	if err != nil {
		t.Fatalf("spec file parse failed?!: %s", err)
	}
	var patches testmark.PatchAccumulator
	defer func() {
		if *testmark.Regen {
			patches.WriteFileWithPatches(doc, file)
		}
	}()

	// Data hunk in this spec file are in "directories" of a test scenario each.
	doc.BuildDirIndex()
	for _, dir := range doc.DirEnt.ChildrenList {
		t.Run(dir.Name, func(t *testing.T) {
			// We always expect the hunk with dag-jose in hex form.
			// Parse the hex (and strip linebreaks, which should be about every 80 chars, though that's not enforced).
			fixtureDataHex := dir.Children["serial.dag-jose.hex"].Hunk.Body
			fixtureDataHex = bytes.ReplaceAll(fixtureDataHex, []byte{'\n'}, []byte{})
			fixtureDataBinary := make([]byte, hex.DecodedLen(len(fixtureDataHex)))
			i, err := hex.Decode(fixtureDataBinary, fixtureDataHex)
			if err != nil {
				t.Fatalf("invalid fixture: %s at position %v", err, i)
			}

			t.Run("decode", func(t *testing.T) {
				n, err := ipld.Decode(fixtureDataBinary, Decode)
				if err != nil {
					t.Fatalf("%s", err)
				}

				t.Run("reencode", func(t *testing.T) {
					reencodeBinary, err := ipld.Encode(n, Encode)
					if err != nil {
						t.Fatalf("%s", err)
					}
					// Encode back to hex string.  We'll diff on this as the test because if it's not equal, that produces the most readable feedback.
					reencodeHex := hex.EncodeToString(reencodeBinary)
					qt.Check(t, reencodeHex, qt.Equals, string(fixtureDataHex))
				})

				if fixtureCid, exists := dir.Children["serial.dag-jose.cid"]; exists {
					t.Run("match-cid", func(t *testing.T) {
						var linkSystem = cidlink.DefaultLinkSystem()
						if lnk, err := linkSystem.ComputeLink(dagJOSELink, n); err != nil {
							t.Fatalf("%s", err)
						} else {
							fixtureCidString := strings.TrimSpace(string(fixtureCid.Hunk.Body))
							qt.Check(t, lnk.String(), qt.Equals, fixtureCidString)
						}
					})
				}

				if fixturePaths, exists := dir.Children["paths"]; exists {
					t.Run("datamodel-pathlist", func(t *testing.T) {
						var foundPaths bytes.Buffer
						traversal.Walk(n, func(tp traversal.Progress, _ datamodel.Node) error {
							if tp.Path.Len() == 0 {
								return nil
							}
							foundPaths.WriteString(tp.Path.String())
							foundPaths.WriteRune('\n')
							return nil
						})

						if *testmark.Regen {
							patches.AppendPatchIfBodyDiffers(*fixturePaths.Hunk, foundPaths.Bytes())
						} else {
							qt.Check(t, foundPaths.String(), qt.Equals, string(fixturePaths.Hunk.Body))
						}
					})
				}

				if fixtureJson, exists := dir.Children["datamodel.dag-json.pretty"]; exists {
					t.Run("datamodel-dagjson", func(t *testing.T) {
						dagjson, err := ipld.Encode(n, dagjson.Encode)
						if err != nil {
							t.Fatalf("%s", err)
						}
						var dagjsonPretty bytes.Buffer
						json.Indent(&dagjsonPretty, dagjson, "", "\t")

						if *testmark.Regen {
							patches.AppendPatchIfBodyDiffers(*fixtureJson.Hunk, dagjsonPretty.Bytes())
						} else {
							qt.Check(t, dagjsonPretty.String()+"\n", qt.Equals, string(fixtureJson.Hunk.Body))
						}
					})
				}
			})
		})
	}
}
