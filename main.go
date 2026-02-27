package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	DefaultFile string `json:"default_file"`
	BaseDir     string `json:"base_dir"`
	OutDir      string `json:"out_dir"`
	DefaultExt  string `json:"default_ext"`
	WikiLang    string `json:"wiki_lang"`
	ProcessTopN int    `json:"process_top_n"`
}

func main() {
	config := loadConfig("config.txt")

	os.MkdirAll(config.OutDir, os.ModePerm)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n===== MENU =====")
		fmt.Println("A - Analyse fichier")
		fmt.Println("B - Analyse multi-fichiers")
		fmt.Println("C - Analyse page Wikipédia")
		fmt.Println("D - ProcessOps")
		fmt.Println("E - SecureOps")
		fmt.Println("Q - Quitter")
		fmt.Print("Choix: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToUpper(choice))

		switch choice {
		case "A":
			analyseSingleFile(config, reader)
		case "B":
			analyseMultiFiles(config, reader)
		case "C":
			cfg := Config{}
			analyseWikipedia(cfg, reader)
		case "D":
			processMenu(reader)
		case "E":
			cfg := Config{}
			secureMenu(cfg, reader)
		case "Q":
			fmt.Println("Fin.")
			return
		default:
			fmt.Println("Choix invalide")
		}
	}
}

func loadConfig(path string) Config {
	cfg := Config{
		DefaultFile: "data/input.txt",
		BaseDir:     "data",
		OutDir:      "out",
		DefaultExt:  ".txt",
	}

	file, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "default_file":
			cfg.DefaultFile = val
		case "base_dir":
			cfg.BaseDir = val
		case "out_dir":
			cfg.OutDir = val
		case "default_ext":
			cfg.DefaultExt = val
		}
	}

	return cfg
}

func analyseSingleFile(cfg Config, reader *bufio.Reader) {
	fmt.Print("Chemin fichier (vide = défaut): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	if path == "" {
		path = cfg.DefaultFile
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		fmt.Println("Fichier invalide.")
		return
	}

	lines := readLines(path)

	// Infos fichier
	fmt.Println("Taille:", info.Size(), "bytes")
	fmt.Println("Modifié:", info.ModTime().Format(time.RFC3339))
	fmt.Println("Nb lignes:", len(lines))

	// Stats mots
	wordCount := 0
	totalLength := 0

	for _, line := range lines {
		words := strings.Fields(line)
		for _, w := range words {
			if _, err := strconv.Atoi(w); err == nil {
				continue
			}
			wordCount++
			totalLength += len(w)
		}
	}

	if wordCount > 0 {
		fmt.Println("Nb mots:", wordCount)
		fmt.Println("Longueur moyenne:", totalLength/wordCount)
	}

	// Mot clé
	fmt.Print("Mot clé: ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)

	var filtered []string
	var notFiltered []string

	for _, line := range lines {
		if strings.Contains(line, keyword) {
			filtered = append(filtered, line)
		} else {
			notFiltered = append(notFiltered, line)
		}
	}

	writeLines(filepath.Join(cfg.OutDir, "filtered.txt"), filtered)
	writeLines(filepath.Join(cfg.OutDir, "filtered_not.txt"), notFiltered)

	// Head / Tail
	fmt.Print("Nombre lignes head/tail: ")
	nStr, _ := reader.ReadString('\n')
	nStr = strings.TrimSpace(nStr)
	n, _ := strconv.Atoi(nStr)

	if n > len(lines) {
		n = len(lines)
	}

	writeLines(filepath.Join(cfg.OutDir, "head.txt"), lines[:n])
	writeLines(filepath.Join(cfg.OutDir, "tail.txt"), lines[len(lines)-n:])

	fmt.Println("Analyse terminée.")
}

func analyseMultiFiles(cfg Config, reader *bufio.Reader) {
	fmt.Print("Répertoire (vide = base_dir): ")
	dir, _ := reader.ReadString('\n')
	dir = strings.TrimSpace(dir)

	if dir == "" {
		dir = cfg.BaseDir
	}

	var report []string
	var index []string
	var merged []string

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, cfg.DefaultExt) {

			report = append(report, fmt.Sprintf("File: %s | Size: %d", path, info.Size()))
			index = append(index, fmt.Sprintf("%s | %d | %s",
				path, info.Size(), info.ModTime().Format(time.RFC3339)))

			lines := readLines(path)
			merged = append(merged, lines...)
		}
		return nil
	})

	writeLines(filepath.Join(cfg.OutDir, "report.txt"), report)
	writeLines(filepath.Join(cfg.OutDir, "index.txt"), index)
	writeLines(filepath.Join(cfg.OutDir, "merged.txt"), merged)

	fmt.Println("Analyse multi-fichiers terminée.")
}

func readLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func writeLines(path string, lines []string) {
	file, _ := os.Create(path)
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
}

func analyseWikipedia(cfg Config, reader *bufio.Reader) {

	fmt.Print("Nom article (ex: Go_(langage)): ")
	article, _ := reader.ReadString('\n')
	article = strings.TrimSpace(article)

	if article == "" {
		fmt.Println("Article invalide")
		return
	}

	url := "https://fr.wikipedia.org/wiki/" + article

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "WebOpsBot/1.0")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Erreur HTTP:", resp.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Erreur parsing HTML:", err)
		return
	}

	var paragraphs []string

	// Extraction des paragraphes <p>
	doc.Find("#mw-content-text p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			paragraphs = append(paragraphs, text)
		}
	})

	if len(paragraphs) == 0 {
		fmt.Println("Aucun paragraphe trouvé.")
		return
	}

	// =========================
	// Traitement 1 : Stats mots
	// =========================

	wordCount := 0
	totalLength := 0

	for _, p := range paragraphs {
		words := strings.Fields(p)
		for _, w := range words {
			if _, err := strconv.Atoi(w); err == nil {
				continue
			}
			wordCount++
			totalLength += len(w)
		}
	}

	avg := 0
	if wordCount > 0 {
		avg = totalLength / wordCount
	}

	// =========================
	// Traitement 2 : Mot-clé
	// =========================

	fmt.Print("Mot clé à filtrer: ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)

	var filtered []string
	for _, p := range paragraphs {
		if strings.Contains(strings.ToLower(p), strings.ToLower(keyword)) {
			filtered = append(filtered, p)
		}
	}

	// =========================
	// Écriture résultat
	// =========================

	outputPath := filepath.Join(cfg.OutDir, "wiki_"+article+".txt")

	var result []string
	result = append(result, "Article: "+article)
	result = append(result, "Total paragraphes: "+strconv.Itoa(len(paragraphs)))
	result = append(result, "Total mots: "+strconv.Itoa(wordCount))
	result = append(result, "Longueur moyenne mot: "+strconv.Itoa(avg))
	result = append(result, "")
	result = append(result, "Paragraphes contenant '"+keyword+"':")
	result = append(result, filtered...)

	writeLines(outputPath, result)

	fmt.Println("Analyse Wikipédia terminée →", outputPath)
}

// Fontion pour afficher le menu du process - Sous-menu ProcessOps

func processMenu(reader *bufio.Reader) {

	for {
		fmt.Println("\n===== ProcessOps =====")
		fmt.Println("1 - Lister processus (top 20)")
		fmt.Println("2 - Rechercher processus")
		fmt.Println("3 - Kill sécurisé")
		fmt.Println("Q - Retour")
		fmt.Print("Choix: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToUpper(choice))

		switch choice {
		case "1":
			listProcesses(20)
		case "2":
			searchProcess(reader)
		case "3":
			killProcess(reader)
		case "Q":
			return
		default:
			fmt.Println("Choix invalide")
		}
	}
}

// Lister processus

func listProcesses(limit int) {

	fmt.Println("OS détecté:", runtime.GOOS)

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/FO", "CSV")
	} else { // macOS
		cmd = exec.Command("ps", "-Ao", "pid,comm")
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Erreur exécution:", err)
		return
	}

	lines := strings.Split(string(output), "\n")

	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fmt.Println(line)
		count++

		if count >= limit {
			break
		}
	}
}

// Rechercher processus

func searchProcess(reader *bufio.Reader) {

	fmt.Print("Mot clé: ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(strings.ToLower(keyword))

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		cmd = exec.Command("ps", "-Ao", "pid,comm")
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Erreur:", err)
		return
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), keyword) {
			fmt.Println(line)
		}
	}
}

// Kill sécurisé

func killProcess(reader *bufio.Reader) {

	fmt.Print("PID à tuer: ")
	pidStr, _ := reader.ReadString('\n')
	pidStr = strings.TrimSpace(pidStr)

	// Vérifier que c'est un nombre
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Println("PID invalide")
		return
	}

	// Affichage confirmation
	fmt.Println("Vous allez tuer le processus PID:", pid)
	fmt.Print("Confirmer (yes/no): ")

	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "yes" {
		fmt.Println("Annulé.")
		return
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/PID", pidStr, "/T")
	} else {
		cmd = exec.Command("kill", pidStr)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Println("Erreur kill:", err)
		return
	}

	fmt.Println("Processus terminé (si autorisé).")
}

// Fonction pour Chargement JSON

func loadConfigJSON(path string) Config {

	// Valeurs par défaut
	cfg := Config{
		DefaultFile: "data/input.txt",
		BaseDir:     "data",
		OutDir:      "out",
		DefaultExt:  ".txt",
		WikiLang:    "fr",
		ProcessTopN: 20,
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Config introuvable, valeurs par défaut utilisées.")
		return cfg
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println("Erreur parsing JSON, valeurs par défaut utilisées.")
		return cfg
	}

	return cfg
}

// Menu de sécurité

func secureMenu(cfg Config, reader *bufio.Reader) {

	for {
		fmt.Println("\n===== SecureOps =====")
		fmt.Println("1 - Verrouiller fichier")
		fmt.Println("2 - Déverrouiller fichier")
		fmt.Println("Q - Retour")
		fmt.Print("Choix: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToUpper(choice))

		switch choice {
		case "1":
			lockFile(cfg, reader)
		case "2":
			unlockFile(cfg, reader)
		case "Q":
			return
		default:
			fmt.Println("Choix invalide")
		}
	}
}

// Verrouiller

func lockFile(cfg Config, reader *bufio.Reader) {

	fmt.Print("Nom du fichier à verrouiller: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		fmt.Println("Nom invalide")
		return
	}

	lockPath := filepath.Join(cfg.OutDir, name+".lock")

	if _, err := os.Stat(lockPath); err == nil {
		fmt.Println("Fichier déjà verrouillé.")
		return
	}

	file, err := os.Create(lockPath)
	if err != nil {
		fmt.Println("Erreur création lock:", err)
		return
	}
	defer file.Close()

	fmt.Println("Fichier verrouillé :", lockPath)
}

// Dévérrouiller

func unlockFile(cfg Config, reader *bufio.Reader) {

	fmt.Print("Nom du fichier à déverrouiller: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	lockPath := filepath.Join(cfg.OutDir, name+".lock")

	err := os.Remove(lockPath)
	if err != nil {
		fmt.Println("Erreur suppression lock:", err)
		return
	}

	fmt.Println("Fichier déverrouillé.")
}
