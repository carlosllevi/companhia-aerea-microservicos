package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/sijms/go-ora/v2"
)

type Passageiro struct {
	ID    int64  `json:"id"`
	Nome  string `json:"nome"`
	CPF   string `json:"cpf"`
	Email string `json:"email"`
}

var db *sql.DB

func main() {
	var err error
	
	dsn := os.Getenv("ORACLE_PASSAGEIROS_URL")
	if dsn == "" {
		log.Fatal("variável ORACLE_PASSAGEIROS_URL não definida")
	}

	db, err = sql.Open("oracle", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("erro ao conectar no Oracle: ", err)
	}
	log.Println("Conectado ao Oracle")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /passageiros", cadastrar)
	mux.HandleFunc("GET /passageiros/{id}", consultar)

	log.Println("Passageiros rodando na porta 8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func cadastrar(w http.ResponseWriter, r *http.Request) {
	var p Passageiro
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var id int64
	_, err := db.Exec(
		`INSERT INTO PASSAGEIROS (NOME, CPF, EMAIL) VALUES (:1, :2, :3) RETURNING ID INTO :4`,
		p.Nome, p.CPF, p.Email, sql.Out{Dest: &id},
	)
	if err != nil {
		http.Error(w, "erro ao salvar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	p.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func consultar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var p Passageiro
	err = db.QueryRow(
		`SELECT ID, NOME, CPF, EMAIL FROM PASSAGEIROS WHERE ID = :1`, id,
	).Scan(&p.ID, &p.Nome, &p.CPF, &p.Email)
	if err == sql.ErrNoRows {
		http.Error(w, "passageiro não encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "erro ao consultar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}