package main

import (
    "errors"
    "fmt"
    "log"
    "time"
    "strconv"

	"github.com/mike-cellini/chirpy/internal/database"

	"github.com/golang-jwt/jwt/v5"
)

const accessIssuer string = "chirpy-access"
const refreshIssuer string = "chirpy-refresh"

func getAccessIssuerAndExpiration(issuedAt time.Time) (issuer string, expiration time.Time) {
    issuer = accessIssuer
    expiration = issuedAt.Add(time.Duration(1 * int(time.Hour)))
    return issuer, expiration
}

func getRefreshIssuerAndExpiration(issuedAt time.Time) (issuer string, expiration time.Time) {
    issuer = refreshIssuer
    expiration = issuedAt.Add(time.Duration(60 * 24 * int(time.Hour)))
    return issuer, expiration
}

func createToken(userId int, getIssuerAndExpiration func(time.Time) (string, time.Time)) string {
    issuedAt := time.Now().UTC()
    issuer, expiresAt := getIssuerAndExpiration(issuedAt)

    var claims = jwt.RegisteredClaims {
        Issuer: issuer,
        IssuedAt: jwt.NewNumericDate(issuedAt),
        ExpiresAt: jwt.NewNumericDate(expiresAt),
        Subject: fmt.Sprint(userId),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    val, err := token.SignedString([]byte(apiCfg.jwtSecret))
    if err != nil {
        log.Printf("Unable to create token: %s", err.Error())
    }

    return val
}

func validateAccessToken(token string) (int, error) {
    claims := jwt.RegisteredClaims {}
    t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
        return []byte(apiCfg.jwtSecret), nil
    })

    if err != nil {
        log.Printf("Token parse error: %s", err.Error())
        return 0, err
    }

    s, err := t.Claims.GetSubject()
    if err != nil {
        return 0, errors.New("Unable to parse claims")
    }

    i, err := t.Claims.GetIssuer()
    if err != nil {
        return 0, errors.New("Unable to parse issuer")
    } else if i != accessIssuer {
        return 0, errors.New("Not an access token")
    }

    id, _ := strconv.Atoi(s)
    return id, nil
}

func validateRefreshToken(token string, db *database.DB) (int, error) {
    claims := jwt.RegisteredClaims {}
    t, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
        return []byte(apiCfg.jwtSecret), nil
    })

    if err != nil {
        log.Printf("Token parse error: %s", err.Error())
        return 0, err
    }

    s, err := t.Claims.GetSubject()
    if err != nil {
        log.Printf("Unable to parse claims")
        return 0, errors.New("Unable to parse claims")
    }

    i, err := t.Claims.GetIssuer()
    if err != nil {
        log.Printf("Unable to parse issuer")
        return 0, errors.New("Unable to parse issuer")
    } else if i != refreshIssuer {
        log.Printf("Not a refresh token")
        return 0, errors.New("Not a refresh token")
    }

    _, ok, err := db.GetRevokedToken(token)
    if err != nil {
        log.Printf("ERROR: %s", err.Error())
        return 0, err
    } else if ok {
        log.Printf("Refresh token revoked")
        return 0, errors.New("Refresh token revoked")
    }
    
    id, _ := strconv.Atoi(s)
    return id, nil
}

func revokeToken(token string, db *database.DB) error {
    _, err := validateRefreshToken(token, db)
    if (err != nil) {
        return err 
    }

    err = db.CreateRevokedToken(token)
    if (err != nil) {
        log.Printf("Unable to create revoked token: %s", err.Error())
        return err 
    }
    return nil
}
