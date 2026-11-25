package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/kjwardy/go-url/api/config"
)

var errSessionNotFound = errors.New("session not found")

type sessionRepository interface {
	load(context.Context, string) ([]byte, error)
	save(context.Context, string, []byte, time.Time) error
	delete(context.Context, string) error
}

type postgresSessionRepository struct {
	database *pg.DB
}

func (r postgresSessionRepository) load(ctx context.Context, id string) ([]byte, error) {
	var record struct {
		Data []byte
	}
	_, err := r.database.WithContext(ctx).QueryOne(
		&record,
		"SELECT data FROM go_url_sessions WHERE id = ? AND expires_at > now()",
		id,
	)
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return record.Data, nil
}

func (r postgresSessionRepository) save(ctx context.Context, id string, data []byte, expiresAt time.Time) error {
	database := r.database.WithContext(ctx)
	if _, err := database.Exec("DELETE FROM go_url_sessions WHERE expires_at <= now()"); err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	_, err := database.Exec(`
		INSERT INTO go_url_sessions (id, data, expires_at)
		VALUES (?, ?, ?)
		ON CONFLICT (id) DO UPDATE
		SET data = EXCLUDED.data, expires_at = EXCLUDED.expires_at
	`, id, data, expiresAt)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (r postgresSessionRepository) delete(ctx context.Context, id string) error {
	_, err := r.database.WithContext(ctx).Exec("DELETE FROM go_url_sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

type postgresSessionStore struct {
	repository sessionRepository
	codecs     []securecookie.Codec
	options    *sessions.Options
}

func newPostgresSessionStore(database *pg.DB, authConfig config.Auth) *postgresSessionStore {
	return newSessionStore(postgresSessionRepository{database: database}, authConfig)
}

func newSessionStore(repository sessionRepository, authConfig config.Auth) *postgresSessionStore {
	cookieStore := sessions.NewCookieStore([]byte(authConfig.SessionToken))
	cookieStore.MaxAge(authConfig.MaxAge)
	return &postgresSessionStore{
		repository: repository,
		codecs:     cookieStore.Codecs,
		options: &sessions.Options{
			Path:     "/",
			MaxAge:   authConfig.MaxAge,
			HttpOnly: true,
			Secure:   authConfig.SecureCookies,
			SameSite: http.SameSiteLaxMode,
		},
	}
}

func (s *postgresSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *postgresSessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	options := *s.options
	session.Options = &options
	session.IsNew = true

	cookie, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}
	var id string
	if err := securecookie.DecodeMulti(name, cookie.Value, &id, s.codecs...); err != nil || id == "" {
		return session, nil
	}
	data, err := s.repository.load(r.Context(), id)
	if errors.Is(err, errSessionNotFound) {
		return session, nil
	}
	if err != nil {
		return session, err
	}
	var values map[interface{}]interface{}
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&values); err != nil {
		return session, fmt.Errorf("decode session: %w", err)
	}
	if values == nil {
		values = make(map[interface{}]interface{})
	}
	session.ID = id
	session.Values = values
	session.IsNew = false
	return session, nil
}

func (s *postgresSessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options == nil {
		return errors.New("session options are missing")
	}
	if session.Options.MaxAge <= 0 {
		if session.ID != "" {
			if err := s.repository.delete(r.Context(), session.ID); err != nil {
				return err
			}
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	id := session.ID
	if id == "" {
		var err error
		id, err = newSessionID()
		if err != nil {
			return err
		}
	}
	encodedID, err := securecookie.EncodeMulti(session.Name(), id, s.codecs...)
	if err != nil {
		return err
	}
	var data bytes.Buffer
	if err := gob.NewEncoder(&data).Encode(session.Values); err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	expiresAt := time.Now().Add(time.Duration(session.Options.MaxAge) * time.Second)
	if err := s.repository.save(r.Context(), id, data.Bytes(), expiresAt); err != nil {
		return err
	}
	session.ID = id
	session.IsNew = false
	http.SetCookie(w, sessions.NewCookie(session.Name(), encodedID, session.Options))
	return nil
}

func newSessionID() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate session ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
