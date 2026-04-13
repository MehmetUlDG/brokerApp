package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey, context değerleri için çakışmayı önleyen özel tip.
type contextKey string

// UserIDKey, context içindeki kullanıcı UUID'sini tutmak için kullanılan anahtar.
const UserIDKey contextKey = "userID"

// JWTAuth, Bearer token doğrulaması yapan HTTP middleware'idir.
//
// Beklenen format: Authorization: Bearer <JWT>
//
// Akış:
//  1. Authorization başlığını kontrol et
//  2. "Bearer <token>" formatını doğrula
//  3. Token imzasını ve son kullanma tarihini kontrol et
//  4. Claims'ten user ID'yi çıkart
//  5. User ID'yi context'e ekle ve handler'a geç
//
// Hatalı/eksik token durumunda 401 Unauthorized döner.
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, "yetkilendirme başlığı eksik")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeUnauthorized(w, "geçersiz format, beklenen: 'Bearer <token>'")
				return
			}

			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				// İmzalama algoritması doğrulama (algorithm confusion saldırısını önler)
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("beklenmeyen imzalama algoritması")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				writeUnauthorized(w, "geçersiz veya süresi dolmuş token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeUnauthorized(w, "token içeriği okunamadı")
				return
			}

			subStr, ok := claims["sub"].(string)
			if !ok || subStr == "" {
				writeUnauthorized(w, "token içinde kullanıcı ID bulunamadı")
				return
			}

			userID, err := uuid.Parse(subStr)
			if err != nil {
				writeUnauthorized(w, "geçersiz kullanıcı ID formatı")
				return
			}

			// User ID'yi context'e ekle ve bir sonraki handler'a ilet
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID, JWT middleware tarafından context'e eklenen kullanıcı ID'sini döner.
// Middleware atlanmışsa veya token geçersizse (false, zero-UUID) döner.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// writeUnauthorized, 401 JSON yanıtı yazar.
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	// Manuel yazım — json.Encoder döngüsel bağımlılık riski yok
	w.Write([]byte(`{"error":"` + message + `","code":"unauthorized"}`)) //nolint:errcheck
}
