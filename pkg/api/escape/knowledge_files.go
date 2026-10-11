package escape

import (
	"context"
	"fmt"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// KnowledgeFileSortTypes lists the sort fields the knowledge file listing
// accepts, mirroring the operation's sortType enum.
var KnowledgeFileSortTypes = []string{"CREATED_AT", "KIND", "NAME", "STATUS"}

// ListKnowledgeFiles lists one page of the organization knowledge library.
func ListKnowledgeFiles(ctx context.Context, cursor string, size int, sortType, sortDirection string) ([]v3.KnowledgeFileSummary, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.KnowledgeFilesAPI.ListKnowledgeFiles(ctx)
	if cursor != "" {
		req = req.Cursor(cursor)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if sortType != "" {
		req = req.SortType(sortType)
	}

	if sortDirection != "" {
		req = req.SortDirection(sortDirection)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.GetData(), data.NextCursor, data.GetTotalCount(), nil
}

// UploadKnowledgeFile registers a file whose bytes were already PUT to a signed
// upload URL, keyed by the id that URL returned.
func UploadKnowledgeFile(ctx context.Context, name, kind, temporaryObjectKey string) (*v3.UploadKnowledgeFile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := v3.UploadKnowledgeFileRequest{
		Name:               name,
		Kind:               v3.ENUMPROPERTIESPENTESTPROPERTIESFILESITEMSPROPERTIESKIND(kind),
		TemporaryObjectKey: temporaryObjectKey,
	}

	data, _, err := client.KnowledgeFilesAPI.UploadKnowledgeFile(ctx).UploadKnowledgeFileRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeleteKnowledgeFile deletes a knowledge file. The generated delete operation
// reuses the {"message": string} schema shared with the other delete routes.
func DeleteKnowledgeFile(ctx context.Context, fileID string) (*v3.DeleteProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.KnowledgeFilesAPI.DeleteKnowledgeFile(ctx, fileID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// knowledgeFileKindFromExtension maps a file extension to the kind the API
// expects. The API derives the kind from the extension, so a caller that names
// a file gets the same value the server would compute.
func knowledgeFileKindFromExtension(name string) (string, bool) {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] != '.' {
			continue
		}

		switch strings.ToLower(name[i+1:]) {
		case "csv":
			return "CSV", true
		case "doc":
			return "DOC", true
		case "docx":
			return "DOCX", true
		case "graphql", "gql":
			return "GRAPHQL", true
		case "html", "htm":
			return "HTML", true
		case "json":
			return "JSON", true
		case "md", "markdown":
			return "MD", true
		case "odt":
			return "ODT", true
		case "pdf":
			return "PDF", true
		case "ppt":
			return "PPT", true
		case "pptx":
			return "PPTX", true
		case "tar", "gz", "tgz":
			return "TAR_GZ", true
		case "txt", "text":
			return "TXT", true
		case "xls":
			return "XLS", true
		case "xlsx":
			return "XLSX", true
		case "xml":
			return "XML", true
		case "zip":
			return "ZIP", true
		}

		return "FILE", true
	}

	return "", false
}

// KnowledgeFileKind resolves the kind of a file name. Unknown extensions map to
// FILE, which is the kind the API uses for anything it cannot classify.
func KnowledgeFileKind(name string) string {
	if kind, ok := knowledgeFileKindFromExtension(name); ok {
		return kind
	}

	return "FILE"
}
