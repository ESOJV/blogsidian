package internal

import (
	"database/sql"
	"encoding/json"
)

type Post struct {
	Slug      string   `yaml:"slug"       json:"slug"`
	Title     string   `yaml:"title"      json:"title"`
	Date      string   `yaml:"date"       json:"date"`
	Tags      []string `yaml:"tags"       json:"tags"`
	Published bool     `yaml:"published"  json:"published"`
	Content   string   `json:"content"`
}

// Creates post or updates on slug conflict
func UpsertPost(db *sql.DB, post *Post) error {
	tags, err := json.Marshal(post.Tags)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO posts(slug, title, date, tags, content)
	VALUES(?, ?, ?, ?, ?)
	ON CONFLICT(slug) DO UPDATE SET
	title = excluded.title,
	date = excluded.date,
	tags = excluded.tags,
	content = excluded.content`,
		post.Slug, post.Title, post.Date, tags, post.Content)

	return err
}

func GetPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query("SELECT slug, title, date, tags, content FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var tagsJSON string
		err := rows.Scan(&p.Slug, &p.Title, &p.Date, &tagsJSON, &p.Content)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(tagsJSON), &p.Tags)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func GetPostBySlug(db *sql.DB, slug string) (*Post, error) {
	row := db.QueryRow(`SELECT slug, title, date, tags,content FROM posts WHERE slug = ?;`, slug)

	var p Post
	var tagsJSON string
	err := row.Scan(&p.Slug, &p.Title, &p.Date, &tagsJSON, &p.Content)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(tagsJSON), &p.Tags)
	if err != nil {
		return nil, err
	}

	return &p, nil

}

func DeletePostBySlug(db *sql.DB, slug string) error {
	_, err := db.Exec(`DELETE FROM posts WHERE slug = (?)`, slug)
	if err != nil {
		return err
	}
	return nil
}
