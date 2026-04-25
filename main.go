package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
)

type Pokemon struct {
	Name    string   `json:"name"`
	HP      int      `json:"hp"`
	MaxHP   int      `json:"max_hp"`
	Attack  int      `json:"attack"`
	Defense int      `json:"defense"`
	Speed   int      `json:"speed"`
	Types   []string `json:"types"`
	Sprite  string   `json:"sprite"`
	Moves   []Move   `json:"moves"`
}

type Move struct {
	Name  string `json:"name"`
	Power int    `json:"power"`
	Type  string `json:"type"`
}

type TurnResult struct {
	Attacker   string  `json:"attacker"`
	Defender   string  `json:"defender"`
	MoveName   string  `json:"move_name"`
	Damage     int     `json:"damage"`
	Multiplier float64 `json:"multiplier"`
	P1HP       int     `json:"p1_hp"`
	P2HP       int     `json:"p2_hp"`
}

type BattleResult struct {
	P1     *Pokemon     `json:"p1"`
	P2     *Pokemon     `json:"p2"`
	Turns  []TurnResult `json:"turns"`
	Winner string       `json:"winner"`
	Error  string       `json:"error,omitempty"`
}

type apiPokemon struct {
	Name    string `json:"name"`
	Sprites struct {
		Front string `json:"front_default"`
	} `json:"sprites"`
	Stats []struct {
		Base int `json:"base_stat"`
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Moves []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
	} `json:"moves"`
}

type apiMove struct {
	Name  string `json:"name"`
	Power int    `json:"power"`
	Type  struct {
		Name string `json:"name"`
	} `json:"type"`
}

type BattleActor interface {
	TakeDamage(amount int)
	IsAlive() bool
	GetName() string
	GetHP() int
}

func (p *Pokemon) TakeDamage(amount int) {
	p.HP -= amount
	if p.HP < 0 {
		p.HP = 0
	}
}

func (p *Pokemon) IsAlive() bool   { return p.HP > 0 }
func (p *Pokemon) GetName() string { return p.Name }
func (p *Pokemon) GetHP() int      { return p.HP }

var typeChart = map[string]map[string]float64{
	"fire":     {"grass": 2, "ice": 2, "bug": 2, "steel": 2, "water": 0.5, "rock": 0.5, "dragon": 0.5},
	"water":    {"fire": 2, "ground": 2, "rock": 2, "water": 0.5, "grass": 0.5, "dragon": 0.5},
	"electric": {"water": 2, "flying": 2, "electric": 0.5, "grass": 0.5, "ground": 0},
	"grass":    {"water": 2, "ground": 2, "rock": 2, "fire": 0.5, "grass": 0.5, "flying": 0.5, "bug": 0.5, "dragon": 0.5},
	"psychic":  {"fighting": 2, "poison": 2, "psychic": 0.5, "dark": 0},
	"ice":      {"grass": 2, "ground": 2, "flying": 2, "dragon": 2, "fire": 0.5, "water": 0.5, "ice": 0.5},
	"fighting": {"normal": 2, "ice": 2, "rock": 2, "dark": 2, "steel": 2, "flying": 0.5, "psychic": 0.5, "ghost": 0},
	"dragon":   {"dragon": 2, "steel": 0.5, "fairy": 0},
	"normal":   {"rock": 0.5, "steel": 0.5, "ghost": 0},
}

func getMultiplier(moveType string, defenderTypes []string) float64 {
	total := 1.0
	chart, ok := typeChart[moveType]
	if !ok {
		return 1.0
	}
	for _, dt := range defenderTypes {
		if m, ok := chart[dt]; ok {
			total *= m
		}
	}
	return total
}

func fetchMove(url string) (Move, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Move{}, err
	}
	defer resp.Body.Close()

	var am apiMove
	if err := json.NewDecoder(resp.Body).Decode(&am); err != nil {
		return Move{}, err
	}
	power := am.Power
	if power <= 0 {
		power = 40
	}
	return Move{Name: am.Name, Power: power, Type: am.Type.Name}, nil
}

func fetchPokemon(name string) (*Pokemon, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + name)
	if err != nil {
		return nil, fmt.Errorf("could not reach PokéAPI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("pokémon '%s' not found — check the spelling", name)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("PokéAPI returned status %d", resp.StatusCode)
	}

	var ap apiPokemon
	if err := json.NewDecoder(resp.Body).Decode(&ap); err != nil {
		return nil, fmt.Errorf("failed to parse data: %v", err)
	}

	p := &Pokemon{Name: ap.Name, Sprite: ap.Sprites.Front}

	for _, s := range ap.Stats {
		switch s.Stat.Name {
		case "hp":
			p.HP = s.Base
		case "attack":
			p.Attack = s.Base
		case "defense":
			p.Defense = s.Base
		case "speed":
			p.Speed = s.Base
		}
	}
	p.MaxHP = p.HP

	for _, t := range ap.Types {
		p.Types = append(p.Types, t.Type.Name)
	}

	count := 0
	for _, m := range ap.Moves {
		if count >= 4 {
			break
		}
		mv, err := fetchMove(m.Move.URL)
		if err != nil {
			continue
		}
		if mv.Power > 0 {
			p.Moves = append(p.Moves, mv)
			count++
		}
	}
	if len(p.Moves) == 0 {
		p.Moves = []Move{{Name: "tackle", Power: 40, Type: "normal"}}
	}

	return p, nil
}

func runBattle(p1, p2 *Pokemon) ([]TurnResult, string) {
	p1HP := p1.HP
	p2HP := p2.HP
	var turns []TurnResult

	for i := 0; i < 30; i++ {
		mv1 := p1.Moves[rand.Intn(len(p1.Moves))]
		mult1 := getMultiplier(mv1.Type, p2.Types)
		dmg1 := int(float64(mv1.Power) * (float64(p1.Attack) / float64(p2.Defense)) * mult1)
		if dmg1 < 1 && mult1 > 0 {
			dmg1 = 1
		}
		p2HP -= dmg1
		if p2HP < 0 {
			p2HP = 0
		}
		turns = append(turns, TurnResult{p1.Name, p2.Name, mv1.Name, dmg1, mult1, p1HP, p2HP})
		if p2HP <= 0 {
			break
		}

		mv2 := p2.Moves[rand.Intn(len(p2.Moves))]
		mult2 := getMultiplier(mv2.Type, p1.Types)
		dmg2 := int(float64(mv2.Power) * (float64(p2.Attack) / float64(p1.Defense)) * mult2)
		if dmg2 < 1 && mult2 > 0 {
			dmg2 = 1
		}
		p1HP -= dmg2
		if p1HP < 0 {
			p1HP = 0
		}
		turns = append(turns, TurnResult{p2.Name, p1.Name, mv2.Name, dmg2, mult2, p1HP, p2HP})
		if p1HP <= 0 {
			break
		}
	}

	p1.HP = p1HP
	p2.HP = p2HP

	winner := p1.Name
	if p1HP <= 0 {
		winner = p2.Name
	} else if p2HP <= 0 {
		winner = p1.Name
	} else if p2HP > p1HP {
		winner = p2.Name
	}

	return turns, winner
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func battleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p1Name := r.URL.Query().Get("p1")
	p2Name := r.URL.Query().Get("p2")

	if p1Name == "" || p2Name == "" {
		json.NewEncoder(w).Encode(BattleResult{Error: "Please enter two Pokémon names."})
		return
	}

	var p1, p2 *Pokemon
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		p1, err1 = fetchPokemon(p1Name)
	}()

	go func() {
		defer wg.Done()
		p2, err2 = fetchPokemon(p2Name)
	}()

	wg.Wait()

	if err1 != nil {
		json.NewEncoder(w).Encode(BattleResult{Error: err1.Error()})
		return
	}
	if err2 != nil {
		json.NewEncoder(w).Encode(BattleResult{Error: err2.Error()})
		return
	}

	turns, winner := runBattle(p1, p2)
	json.NewEncoder(w).Encode(BattleResult{P1: p1, P2: p2, Turns: turns, Winner: winner})
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/PokemonBackground.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "PokemonBackground.png")
	})
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "styles.css")
	})
	http.HandleFunc("/battle", battleHandler)
	fmt.Println("Server running → http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
