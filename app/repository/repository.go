package repository

import (
	"jit/app"
	"jit/app/index"
	"path"
)

type Repository struct {
	gitPath   string
	workspace app.Workspace
	database  app.Database
	index     *index.Index
	refs      app.Ref
}

func NewRepository(rootPath string) Repository {
	gitPath := path.Join(rootPath, ".git")
	dbPath := path.Join(gitPath, "objects")
	indexPath := path.Join(gitPath, "index")

	workspace := app.NewWorkspace(rootPath)
	database := app.NewDatabase(dbPath)
	idx := index.NewIndex(indexPath)
	refs := app.NewRef(gitPath)

	return Repository{
		gitPath:   gitPath,
		workspace: workspace,
		database:  database,
		index:     idx,
		refs:      refs,
	}
}

func (r *Repository) GetIndex() *index.Index {
	return r.index
}

func (r *Repository) GetDatabase() app.Database {
	return r.database
}

func (r *Repository) GetWorkspace() app.Workspace {
	return r.workspace
}

func (r Repository) GetRefs() app.Ref {
	return r.refs
}
