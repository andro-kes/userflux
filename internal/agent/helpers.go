package agent

import (
	cr "crypto/rand"
	"math"
	"math/rand/v2"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateBody(fields []string) (map[string]string, error) {
	m := make(map[string]string)
	for i := 0; i < len(fields); i++ {
		s, err := generateRandString()
		if err != nil {
			return nil, err
		}
		m[fields[i]] = s
	}

	return m, nil
}

func generateRandString() (string, error) {
	b := make([]byte, 10)
	if _, err := cr.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b), nil
}

// RandomDelay возвращает задержку в зависимости от режима.
// mode: "uniform", "exp", "normal"
// base: среднее/базовое время
// jitterFraction: для "uniform" -> доля (0.3 = ±30%); для "normal" -> sd (время представленное как float64(base) * jitterFraction обычно)
// min, max: ограничение (поставьте 0 для отключения)
func RandomDelay(base time.Duration, mode string, jitterFraction float64, min, max time.Duration) time.Duration {
	var d time.Duration

	switch mode {
	case "uniform":
		// delta = (rand*2-1) * jitterFraction * base
		delta := (rand.Float64()*2.0 - 1.0) * jitterFraction * float64(base)
		d = base + time.Duration(delta)

	case "exp":
		// rand.ExpFloat64() имеет mean=1, масштабируем на base
		d = time.Duration(rand.ExpFloat64() * float64(base))

	case "normal":
		// Box-Muller: z * sd + mean
		u1 := rand.Float64()
		if u1 <= 0 {
			u1 = math.SmallestNonzeroFloat64
		}
		u2 := rand.Float64()
		z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		// интерпретация jitterFraction: доля от base как SD, т.е. sd = jitterFraction * base
		sd := jitterFraction * float64(base)
		d = time.Duration(float64(base) + z*sd)

	default:
		d = base
	}

	// clamp
	if d < 0 {
		d = 0
	}
	if min > 0 && d < min {
		d = min
	}
	if max > 0 && d > max {
		d = max
	}
	return d
}
