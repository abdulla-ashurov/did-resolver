package types

import (
	cheqd "github.com/cheqd/cheqd-node/x/cheqd/types"
	resource "github.com/cheqd/cheqd-node/x/resource/types"
)

type ResolutionDidDocMetadata struct {
	Created     string            `json:"created,omitempty"`
	Updated     string            `json:"updated,omitempty"`
	Deactivated bool              `json:"deactivated,omitempty"`
	VersionId   string            `json:"versionId,omitempty"`
	Resources   []ResourcePreview `json:"linkedResourceMetadata,omitempty"`
}

type ResourcePreview struct {
	ResourceURI       string `json:"resourceURI"`
	CollectionId      string `json:"resourceCollectionId"`
	Name              string `json:"resourceName"`
	ResourceType      string `json:"resourceType"`
	MediaType         string `json:"mediaType"`
	Created           string `json:"created"`
	Checksum          string `json:"checksum"`
	PreviousVersionId string `json:"previousVersionId"`
	NextVersionId     string `json:"nextVersionId"`
}

func NewResolutionDidDocMetadata(did string, metadata cheqd.Metadata, resources []*resource.ResourceHeader) ResolutionDidDocMetadata {
	newMetadata := ResolutionDidDocMetadata{
		metadata.Created,
		metadata.Updated,
		metadata.Deactivated,
		metadata.VersionId,
		[]ResourcePreview(nil),
	}
	if metadata.Resources == nil {
		return newMetadata
	}
	for _, r := range resources {
		resourcePreview := ResourcePreview{
			did + RESOURCE_PATH + r.Id,
			r.CollectionId,
			r.Name,
			r.ResourceType,
			r.MediaType,
			r.Created,
			FixResourceChecksum(r.Checksum),
			r.PreviousVersionId,
			r.NextVersionId,
		}
		newMetadata.Resources = append(newMetadata.Resources, resourcePreview)
	}
	return newMetadata
}

func TransformToFragmentMetadata(metadata ResolutionDidDocMetadata) ResolutionDidDocMetadata {
	metadata.Resources = nil
	return metadata
}
