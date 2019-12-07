package document

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/document"
	
	"opendev.org/airship/airshipctl/pkg/util"
)

type PullOptions struct {
	ForceReset bool
}

// NewDocumentPullCommand creates a new command for generating secret information
func NewDocumentPullCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	po := &PullOptions{}
	masterPassphraseCmd := &cobra.Command{
		Use: "pull",
		Short: "pulls documents from defined repositories",
		Run: func(cmd *cobra.Command, args []string) {
			PullDocs(rootSettings, po)
		},
	}

	flags := masterPassphraseCmd.PersistentFlags()

	flags.BoolVar(&po.ForceReset, "debug", false, "enable verbose output")
	return masterPassphraseCmd
}

func PullDocs(rootSettings *environment.AirshipCTLSettings, po *PullOptions) {

	fpath := filepath.Join("pkg/document/testdata", "repo.yaml")

	repoOptions := &document.RepositorySpec{}
	
	err := util.ReadYAMLFile(fpath, repoOptions)
	if err != nil {
		fmt.Printf("we got an error during unmarshaling: %v\n", err)
	}

	baseDir := "/Users/kkalinovskiy/"

	repo := document.NewRepositoryFromSpec(baseDir, repoOptions)

	fmt.Printf("Downloading repository\n")
	err = repo.Download(po.ForceReset)

	if err != nil {
		fmt.Printf("Error downloading repository: %v\n", err) 
	}

}