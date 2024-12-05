package dbrepo

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

// its a good practice to always timeout DB connection sessions
const dbTimeout = time.Second * 3

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) AllMovies(genre ...int) ([]*models.Movie, error) {
	// NOTES: Here is how u set a timeout on a db connection session
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("WHERE id IN (SELECT movie_id FROM movies_genres WHERE genre_id = %d)", genre[0])
	}

	query := fmt.Sprintf(`
		SELECT id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		FROM movies %s
		ORDER BY title
	`, where)

	rows, err := m.DB.QueryContext(context, query)
	if err != nil {
		return nil, err
	}
	// NOTES: always close rows
	defer rows.Close()

	var movies []*models.Movie

	for rows.Next() {
		var movie models.Movie
		// NOTES: when u scan returned query recs, u must scan em in same order as they are in the query
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.RunTime,
			&movie.MPAARating,
			&movie.Description,
			&movie.Image,
			&movie.CreatedAt,
			&movie.UpdatedField,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *PostgresDBRepo) OneMovie(id int) (*models.Movie, error) {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		FROM movies
		WHERE id = $1
	`
	row := m.DB.QueryRowContext(context, query, id)
	var movie models.Movie

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedField,
	)
	if err != nil {
		return nil, err
	}

	// get the genres of the selected movie if any
	query = `
		SELECT g.id, g.genre
		FROM movies_genres mg
		LEFT JOIN genres g 
		ON (mg.genre_id = g.id)
		WHERE mg.movie_id = $1
		ORDER BY g.genre
	`
	rows, err := m.DB.QueryContext(context, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	movie.Genres = genres

	return &movie, err
}

func (m *PostgresDBRepo) OneMovieForEdit(id int) (*models.Movie, []*models.Genre, error) {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		FROM movies
		WHERE id = $1
	`
	row := m.DB.QueryRowContext(context, query, id)
	var movie models.Movie

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedField,
	)
	if err != nil {
		return nil, nil, err
	}

	// get the genres of the selected movie if any
	query = `
		SELECT g.id, g.genre
		FROM movies_genres mg
		LEFT JOIN genres g 
		ON (mg.genre_id = g.id)
		WHERE mg.movie_id = $1
		ORDER BY g.genre
	`
	rows, err := m.DB.QueryContext(context, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	var genresArray []int

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}

		genres = append(genres, &g)
		genresArray = append(genresArray, g.ID)
	}

	movie.Genres = genres
	movie.GenresArray = genresArray

	var allGenres []*models.Genre

	query = "SELECT id, genre FROM genres ORDER BY genre"
	gRows, err := m.DB.QueryContext(context, query)
	if err != nil {
		return nil, nil, err
	}
	defer gRows.Close()
	for gRows.Next() {
		var g models.Genre
		err := gRows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}

		allGenres = append(allGenres, &g)
	}

	return &movie, allGenres, err
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, first_name, last_name, password, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user models.User
	row := m.DB.QueryRowContext(context, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *PostgresDBRepo) GetUserById(id int) (*models.User, error) {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, first_name, last_name, password, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var user models.User
	row := m.DB.QueryRowContext(context, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *PostgresDBRepo) AllGenres() ([]*models.Genre, error) {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, genre, created_at, updated_at
		FROM genres 
		ORDER BY genre
	`
	rows, err := m.DB.QueryContext(context, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
			&g.CreatedAt,
			&g.UpdatedField,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	return genres, nil
}

func (m *PostgresDBRepo) InsertMovie(movie models.Movie) (int, error) {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `INSERT INTO movies (title, description, release_date, runtime, mpaa_rating, created_at, updated_at, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id`

	var newID int

	err := m.DB.QueryRowContext(context, stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.RunTime,
		movie.MPAARating,
		movie.CreatedAt,
		movie.UpdatedField,
		movie.Image,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *PostgresDBRepo) UpdateMovie(movie models.Movie) error {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		UPDATE movies SET title = $1,
		description = $2,
		release_date = $3,
		runtime = $4,
		mpaa_rating = $5,
		updated_at = $6,
		image = $7
		WHERE id = $8`

	_, err := m.DB.ExecContext(context, stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.RunTime,
		movie.MPAARating,
		movie.UpdatedField,
		movie.Image,
		movie.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *PostgresDBRepo) UpdateMovieGenres(id int, genreIDs []int) error {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM movies_genres WHERE movie_id = $1`

	_, err := m.DB.ExecContext(context, stmt, id)
	if err != nil {
		return err
	}

	for _, gi := range genreIDs {
		stmt := `INSERT INTO movies_genres (movie_id, genre_id) VALUES ($1, $2)`
		_, err := m.DB.ExecContext(context, stmt, id, gi)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *PostgresDBRepo) DeleteMovie(id int) error {
	context, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM movies WHERE id = $1`

	_, err := m.DB.ExecContext(context, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
