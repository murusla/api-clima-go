package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var apiKey = os.Getenv("OPENWEATHER_KEY")

func main() {
	http.HandleFunc("/clima/", func(w http.ResponseWriter, r *http.Request) {
		cidades := strings.Split(strings.TrimPrefix(r.URL.Path, "/clima/"), ",")

		var wg sync.WaitGroup
		resultados := make([]map[string]interface{}, 0)

		for _, cidade := range cidades {
			wg.Add(1)
			go func(cidade string) {
				defer wg.Done()

				u := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=pt_br",
					url.QueryEscape(cidade), apiKey)

				resp, err := http.Get(u)
				if err != nil || resp == nil {
					return
				}
				defer resp.Body.Close()

				var dados map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&dados)

				// Verifica se a resposta tem os campos esperados
				mainData, ok1 := dados["main"].(map[string]interface{})
				weatherList, ok2 := dados["weather"].([]interface{})

				if !ok1 || !ok2 || len(weatherList) == 0 {
					return // pula cidade inválida ou resposta sem dados
				}

				weather := weatherList[0].(map[string]interface{})

				resultado := map[string]interface{}{
					"cidade":      dados["name"],
					"temperatura": fmt.Sprintf("%.1f°C", mainData["temp"]),
					"condicao":    weather["description"],
				}

				// usa mutex pra adicionar ao slice com segurança
				mutex.Lock()
				resultados = append(resultados, resultado)
				mutex.Unlock()
			}(cidade)
		}

		wg.Wait()
		json.NewEncoder(w).Encode(resultados)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Servidor rodando na porta " + port)
	http.ListenAndServe(":"+port, nil)
}

var mutex sync.Mutex
