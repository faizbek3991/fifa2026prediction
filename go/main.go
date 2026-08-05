package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

// --- Data Structures ---
type SimRequest struct {
	NumSims int                `json:"num_sims"`
	Elos    map[string]float64 `json:"elos"`
}

type TeamResult struct {
	Team        string  `json:"team"`
	Probability float64 `json:"probability"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Simulation Logic ---
func getWinProbability(ratingA, ratingB float64) float64 {
	return 1.0 / (1.0 + math.Pow(10, (ratingB-ratingA)/400.0))
}

func simulateMatch(teamA, teamB string, elos map[string]float64, r *rand.Rand) string {
	probA := getWinProbability(elos[teamA], elos[teamB])
	if r.Float64() < probA {
		return teamA
	}
	return teamB
}

func simulateTournament(elos map[string]float64, r *rand.Rand) string {
	// Quarter-finals
	qf2 := simulateMatch("Spain", "Belgium", elos, r)
	qf3 := simulateMatch("Norway", "England", elos, r)
	qf4 := simulateMatch("Argentina", "Switzerland", elos, r)

	// Semi-finals
	sf1 := simulateMatch("France", qf2, elos, r)
	sf2 := simulateMatch(qf3, qf4, elos, r)

	// Final
	return simulateMatch(sf1, sf2, elos, r)
}

// --- The Go Secret Weapon: Concurrent Simulation ---
func runSimulations(numSims int, elos map[string]float64) map[string]int {
	results := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	simsPerWorker := numSims / numWorkers

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		src := rand.NewSource(time.Now().UnixNano() + int64(i))
		r := rand.New(src)

		go func(count int, localRand *rand.Rand) {
			defer wg.Done()
			localResults := make(map[string]int)

			for j := 0; j < count; j++ {
				winner := simulateTournament(elos, localRand)
				localResults[winner]++
			}

			mu.Lock()
			for k, v := range localResults {
				results[k] += v
			}
			mu.Unlock()
		}(simsPerWorker, r)
	}
	wg.Wait()
	return results
}

func main() {
	seedUsername := flag.String("seed-admin-username", "", "Create/update an admin user with this username and exit")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "users.db"
	}
	db, err := openDB(dbPath)
	if err != nil {
		slog.Error("Failed to open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if *seedUsername != "" {
		password := os.Getenv("SEED_PASSWORD")
		if password == "" {
			fmt.Fprintln(os.Stderr, "SEED_PASSWORD env var must be set when using -seed-admin-username")
			os.Exit(1)
		}
		hash, err := hashPassword(password)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to hash password:", err)
			os.Exit(1)
		}
		if err := createUser(db, *seedUsername, hash); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create user:", err)
			os.Exit(1)
		}
		fmt.Printf("User %q created.\n", *seedUsername)
		return
	}

	limiter := newRateLimiter(5, time.Minute)

	// Serve login page at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "login.html")
	})

	// Serve dashboard after login
	http.HandleFunc("/dashboard", requireAuth(db, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	}))

	// Logout API endpoint
	http.HandleFunc("/api/logout", requireAuth(db, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Warn("Invalid request method on /api/logout", slog.String("method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if cookie, err := r.Cookie(sessionCookieName); err == nil {
			_ = deleteSession(db, cookie.Value)
		}
		clearSessionCookie(w, r)
		slog.Info("User logged out", slog.String("username", usernameFromContext(r.Context())), slog.String("ip", clientIP(r)))
		response := map[string]string{"message": "Logged out successfully"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Login API endpoint
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Warn("Invalid request method on /api/login", slog.String("method", r.Method), slog.String("ip", clientIP(r)))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ip := clientIP(r)
		if !limiter.allow(ip) {
			slog.Warn("Login rate limit exceeded", slog.String("ip", ip))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(LoginResponse{Success: false, Message: "Too many attempts. Try again later."})
			return
		}

		var loginReq LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			slog.Error("Failed to decode login request", slog.Any("error", err), slog.String("ip", ip))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		slog.Info("Login attempt", slog.String("username", loginReq.Username), slog.String("ip", ip))

		hash, err := getPasswordHash(db, loginReq.Username)
		userExists := err == nil
		if !userExists {
			hash = dummyHash // keep bcrypt cost constant so response time can't reveal whether the username exists
		}
		valid := checkPassword(hash, loginReq.Password) && userExists

		if valid {
			token, err := newSessionToken()
			if err != nil || createSession(db, token, loginReq.Username, sessionTTL) != nil {
				slog.Error("Failed to create session", slog.Any("error", err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			setSessionCookie(w, r, token)
			logAuthEvent("Successful login", loginReq.Username, ip, true)
			response := LoginResponse{Success: true, Message: "Login successful"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			logAuthEvent("Failed login attempt", loginReq.Username, ip, false)
			response := LoginResponse{Success: false, Message: "Invalid username or password"}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
		}
	})

	// API endpoint for the simulation
	http.HandleFunc("/api/simulate", requireAuth(db, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Warn("Invalid request method", slog.String("method", r.Method), slog.String("ip", clientIP(r)))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request", slog.Any("error", err))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if req.NumSims <= 0 || req.NumSims > 1_000_000 {
			http.Error(w, "num_sims must be between 1 and 1,000,000", http.StatusBadRequest)
			return
		}

		slog.Info("Starting simulation", slog.Int("num_sims", req.NumSims), slog.String("ip", clientIP(r)))
		start := time.Now()
		rawResults := runSimulations(req.NumSims, req.Elos)
		duration := time.Since(start)

		var finalResults []TeamResult
		for team, count := range rawResults {
			prob := (float64(count) / float64(req.NumSims)) * 100.0
			finalResults = append(finalResults, TeamResult{Team: team, Probability: prob})
		}

		response := map[string]interface{}{
			"results":        finalResults,
			"duration_ms":    duration.Milliseconds(),
			"cpu_cores_used": runtime.NumCPU(),
		}

		slog.Info("Simulation completed", slog.Int("duration_ms", int(duration.Milliseconds())), slog.Int("cpu_cores", runtime.NumCPU()))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Render (and most PaaS hosts) assign a port via $PORT and require the
	// app to bind to it; ADDR is kept as a fallback for local/VPS use.
	addr := os.Getenv("ADDR")
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	if addr == "" {
		addr = ":8080"
	}
	slog.Info("Server started", slog.String("addr", addr))
	slog.Error("Server stopped", slog.Any("error", http.ListenAndServe(addr, nil)))
}
