# Go CLI DevOps / WebOps / SecureOps Tool

Mini outil CLI en Go pour l’analyse de fichiers, gestion des processus, scraping Wikipédia et sécurisation de fichiers.  
Cross-platform : Windows / macOS.

---

## ⚡ Fonctionnalités

### 1️⃣ FileOps
- Analyse de fichiers texte
  - Taille, nombre de lignes, nombre de mots
  - Longueur moyenne des mots
  - Filtrage par mot-clé
  - Head / Tail : N premières / dernières lignes
- Analyse multi-fichiers dans un dossier
  - Rapport global : `out/report.txt`
  - Index : `out/index.txt`
  - Fusion de fichiers : `out/merged.txt`

### 2️⃣ WikiOps
- Analyse d’articles Wikipédia
- Extraction des paragraphes avec `goquery`
- Stats mots, filtrage par mot-clé
- Sortie : `out/wiki_<article>.txt`
- Possibilité de traiter plusieurs articles

### 3️⃣ ProcessOps
- Lister les processus (top N)
  - Windows : `tasklist`
  - macOS : `ps -Ao pid,comm`
- Rechercher / filtrer un processus par mot-clé
- Kill sécurisé avec confirmation
- Cross-platform avec détection `runtime.GOOS`

### 4️⃣ SecureOps
- Gestion des droits / verrouillage de fichiers
- Lockfile portable : `out/<nom>.lock`
- Déverrouillage
- Vérification d’existence du lock avant écriture

### 5️⃣ Configuration
- JSON : `config.json`
```json
{
  "default_file": "data/input.txt",
  "base_dir": "data",
  "out_dir": "out",
  "default_ext": ".txt",
  "wiki_lang": "fr",
  "process_top_n": 10
}
```

## Installation

```bash
git clone <repo>
cd <repo>
go mod download
go mod tidy
```

## Lancer le programme avec configuration par défaut

```bash
go run .
```

## Structure projet

```
.
├── main.go
├── config.json
├── go.mod
├── go.sum
├── out/                  # fichiers générés
├── data/                 # fichiers d’entrée
├── .gitignore
└── README.md
```