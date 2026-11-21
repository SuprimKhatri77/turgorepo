package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
)

// fakeRepo embeds the interface, so a call the test didn't expect panics.
type fakeRepo struct {
	Store

	user        db.User
	getUserErr  error
	lookedUp    string
	created     *db.CreateUserParams
	createErr   error
	sessionErr  error
	sessions    []db.Session // returned in order by GetActiveSessionByID
	rotateErr   error
	rotated     bool
	revoked     []pgtype.UUID
	sessionsNew int
	newSessErr  error

	revokedUsers []pgtype.UUID
	revokeAllErr error

	verificationToken *db.CreateEmailVerificationTokenParams
	verificationErr   error
	verifiedHash      string
	verifyErr         error
}

func (r *fakeRepo) CreateEmailVerificationToken(_ context.Context, params db.CreateEmailVerificationTokenParams) error {
	r.verificationToken = &params
	return r.verificationErr
}

func (r *fakeRepo) VerifyEmailWithToken(_ context.Context, tokenHash string) (db.User, error) {
	r.verifiedHash = tokenHash
	return r.user, r.verifyErr
}

func (r *fakeRepo) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	r.lookedUp = email
	return r.user, r.getUserErr
}

func (r *fakeRepo) GetUserByID(context.Context, pgtype.UUID) (db.User, error) {
	return r.user, r.getUserErr
}

func (r *fakeRepo) CreateUser(_ context.Context, params db.CreateUserParams) (db.User, error) {
	r.created = &params
	if r.createErr != nil {
		return db.User{}, r.createErr
	}
	return db.User{ID: testUUID(1), Name: params.Name, Email: params.Email, Role: params.Role}, nil
}

func (r *fakeRepo) CreateSession(context.Context, db.CreateSessionParams) (db.Session, error) {
	r.sessionsNew++
	return db.Session{}, r.newSessErr
}

func (r *fakeRepo) GetActiveSessionByID(context.Context, pgtype.UUID) (db.Session, error) {
	if r.sessionErr != nil {
		return db.Session{}, r.sessionErr
	}
	if len(r.sessions) == 0 {
		return db.Session{}, pgx.ErrNoRows
	}
	s := r.sessions[0]
	r.sessions = r.sessions[1:]
	return s, nil
}

func (r *fakeRepo) RotateSessionToken(context.Context, db.RotateSessionTokenParams) (db.Session, error) {
	r.rotated = true
	return db.Session{}, r.rotateErr
}

func (r *fakeRepo) RevokeSession(_ context.Context, id pgtype.UUID) error {
	r.revoked = append(r.revoked, id)
	return nil
}

func (r *fakeRepo) RevokeAllUserSessions(_ context.Context, userID pgtype.UUID) error {
	r.revokedUsers = append(r.revokedUsers, userID)
	return r.revokeAllErr
}

// resetRepo adds the password reset queries to fakeRepo.
type resetRepo struct {
	fakeRepo

	createdToken db.CreatePasswordResetTokenParams
	createErr    error
	resetParams  db.ResetPasswordWithTokenParams
	resetErr     error
}

func (r *resetRepo) CreatePasswordResetToken(_ context.Context, params db.CreatePasswordResetTokenParams) error {
	r.createdToken = params
	return r.createErr
}

func (r *resetRepo) ResetPasswordWithToken(_ context.Context, params db.ResetPasswordWithTokenParams) (db.User, error) {
	r.resetParams = params
	return r.user, r.resetErr
}

type fakeLimiter struct {
	blocked bool
	keys    []string
	cleared []string
}

func (l *fakeLimiter) Clear(_ context.Context, key string) error {
	l.cleared = append(l.cleared, key)
	return nil
}

func (l *fakeLimiter) Allow(_ context.Context, key string, _ int, _ time.Duration) (bool, time.Duration, error) {
	l.keys = append(l.keys, key)
	return !l.blocked, time.Minute, nil
}

type fakeMailer struct {
	sent []email.Message
	err  error
}

func (m *fakeMailer) Send(_ context.Context, msg email.Message) error {
	m.sent = append(m.sent, msg)
	return m.err
}

// inlineJobs runs background jobs immediately, so tests can check their effects.
type inlineJobs struct{}

func (inlineJobs) Go(_ string, job func(ctx context.Context)) { job(context.Background()) }

// fakeRevoker records the cutoffs and session revocations it's asked to set.
type fakeRevoker struct {
	userIDs  []string
	cutoff   time.Time
	ttl      time.Duration
	sessions []string
	err      error
}

func (r *fakeRevoker) RevokeTokensBefore(_ context.Context, userID string, cutoff time.Time, ttl time.Duration) error {
	r.userIDs = append(r.userIDs, userID)
	r.cutoff = cutoff
	r.ttl = ttl
	return r.err
}

func (r *fakeRevoker) RevokeSession(_ context.Context, sessionID string, ttl time.Duration) error {
	r.sessions = append(r.sessions, sessionID)
	r.ttl = ttl
	return r.err
}

func (r *fakeRevoker) Revocations(context.Context, string, string) (session.Revocations, error) {
	return session.Revocations{}, nil
}
