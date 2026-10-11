package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

// knowledgeFileUploadTimeout bounds the signed-URL PUT plus the registration
// call. A large knowledge file over a slow link is the normal case, so the
// bound is generous but finite.
const knowledgeFileUploadTimeout = 15 * time.Minute

var (
	knowledgeFilesSortType      string
	knowledgeFilesSortDirection string
	knowledgeFileListPage       pageFlags

	knowledgeFileUploadName string
	knowledgeFileUploadFile string
	knowledgeFileUploadKind string
)

var knowledgeFilesCmd = &cobra.Command{
	Use:     "knowledge-files",
	Aliases: []string{"knowledge-file", "kf"},
	Short:   "Manage the organization knowledge library (Beta)",
	Long: `Manage Knowledge Files - Context for AI Pentests (Beta)

Files in the organization knowledge library are used as context by AI
pentests. Upload a document once and every later AI pentest can read it.

BETA: the knowledge file library has no release note and is not GA. The
upload flow and the processing status it reports can still change.

COMMON WORKFLOWS:
  • Upload a document:  escape-cli knowledge-files upload --file spec.pdf
  • List the library:   escape-cli knowledge-files list`,
}

var knowledgeFilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the organization knowledge files (Beta)",
	Long: `List the files of the organization knowledge library.

Each file is processed asynchronously after the upload: read the STATUS
column, or poll this listing, until it is COMPLETED or FAILED.`,
	Example: `  escape-cli knowledge-files list
  escape-cli knowledge-files list --sort-by NAME`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.KnowledgeFileSummary{}) {
			return nil
		}

		if err := runPagedList(cmd, knowledgeFileListPage, func(ctx context.Context, cursor string, size int) ([]v3.KnowledgeFileSummary, *string, int, error) {
			return escape.ListKnowledgeFiles(ctx, cursor, size, knowledgeFilesSortType, knowledgeFilesSortDirection)
		}, func(files []v3.KnowledgeFileSummary) []string {
			res := []string{"ID\tNAME\tKIND\tSTATUS\tSIZE\tCREATED AT"}
			for _, file := range files {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
					file.GetId(),
					file.GetName(),
					file.GetKind(),
					file.GetStatus(),
					strconv.Itoa(file.GetSizeInBytes()),
					out.GetShortDate(file.GetCreatedAt().String()),
				))
			}

			return res
		}); err != nil {
			return fmt.Errorf("unable to list knowledge files: %w", err)
		}

		return nil
	},
}

var knowledgeFilesUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to the knowledge library (Beta)",
	Long: `Upload a file to the organization knowledge library.

The upload is the three-step flow the API describes, collapsed into one
command: the CLI asks for a signed upload URL, PUTs the file content to it,
then registers the file with the id that URL returned. Processing starts
after the upload; poll 'knowledge-files list' until the status is COMPLETED or FAILED (inspect a FAILED upload before retrying)
or FAILED.

File names must be unique within the organization: names that sanitize to the
same storage key share a single object, so uploading a new file overwrites
the previous one.`,
	Example: `  escape-cli knowledge-files upload --file architecture.md
  cat spec.pdf | escape-cli knowledge-files upload --name spec.pdf`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.UploadKnowledgeFile200Response{}) {
			return nil
		}

		var data []byte
		var err error
		if knowledgeFileUploadFile != "" {
			data, err = os.ReadFile(knowledgeFileUploadFile)
			if err != nil {
				return fmt.Errorf("unable to read %s: %w", knowledgeFileUploadFile, err)
			}
		} else {
			data, err = io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		}

		if len(data) == 0 {
			return errors.New("no file content provided: pass --file or pipe the file via stdin")
		}

		name := knowledgeFileUploadName
		if name == "" {
			if knowledgeFileUploadFile == "" {
				return errors.New("--name is required when the file is piped on stdin")
			}

			// The stored name is what every reader of the library sees and what
			// the API derives the storage key from, so the local path must not
			// leak into either.
			name = filepath.Base(knowledgeFileUploadFile)
		}

		kind := knowledgeFileUploadKind
		if kind == "" {
			kind = escape.KnowledgeFileKind(name)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), knowledgeFileUploadTimeout)
		defer cancel()

		upload, err := escape.GetUploadSignedURL(ctx)
		if err != nil {
			return fmt.Errorf("unable to get signed url: %w", err)
		}

		if err := escape.UploadFileContent(ctx, upload.GetUrl(), name, data); err != nil {
			return fmt.Errorf("failed to upload %s: %w", name, err)
		}

		file, err := escape.UploadKnowledgeFile(ctx, name, kind, upload.GetId())
		if err != nil {
			return fmt.Errorf("failed to register %s: %w", name, err)
		}

		out.Table(file, func() []string {
			return []string{
				"ID\tNAME\tKIND\tSTATUS",
				fmt.Sprintf("%s\t%s\t%s\t%s", file.GetId(), file.GetName(), file.GetKind(), file.GetStatus()),
			}
		})
		out.Log(name + " uploaded: poll 'escape-cli knowledge-files list' until the status is COMPLETED")

		return nil
	},
}

var knowledgeFilesDeleteCmd = &cobra.Command{
	Use:     "delete file-id",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a knowledge file (Beta)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.DeleteProfile200Response{}) {
			return nil
		}

		proceed, err := confirmDestructive(cmd, "delete knowledge file "+args[0], nil)
		if err != nil || !proceed {
			return err
		}

		if _, err := escape.DeleteKnowledgeFile(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete knowledge file: %w", err)
		}

		out.Log(fmt.Sprintf("Knowledge file %s deleted", args[0]))

		return nil
	},
}

func init() {
	knowledgeFilesCmd.AddCommand(knowledgeFilesListCmd, knowledgeFilesUploadCmd, knowledgeFilesDeleteCmd)

	knowledgeFileListPage.bind(knowledgeFilesListCmd)
	knowledgeFilesListCmd.Flags().StringVar(&knowledgeFilesSortType, "sort-by", "", "sort field: "+strings.Join(escape.KnowledgeFileSortTypes, ", "))
	knowledgeFilesListCmd.Flags().StringVar(&knowledgeFilesSortDirection, "sort-direction", "", "sort direction: asc, desc")

	knowledgeFilesUploadCmd.Flags().StringVar(&knowledgeFileUploadFile, "file", "", "path to the file to upload (reads stdin when omitted)")
	knowledgeFilesUploadCmd.Flags().StringVar(&knowledgeFileUploadName, "name", "", "stored file name including its extension (defaults to the base name of --file)")
	knowledgeFilesUploadCmd.Flags().StringVar(&knowledgeFileUploadKind, "kind", "", "file kind (derived from the extension when omitted)")
	bindDestructiveGate(knowledgeFilesDeleteCmd, true)

	rootCmd.AddCommand(knowledgeFilesCmd)
}
