package dag

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func DeleteTemporaryArtifactContents(ctx context.Context, dag WorkflowDag) {
	for _, artf := range dag.Artifacts() {
		if !artf.ShouldPersistContent() {
			err := artf.DeleteContent(ctx)
			log.Errorf("error deleting temporary artifact result. Artf: %s, Err: %v", artf.Name(), err)
		}
	}
}
