package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/salhe123/forge/internal/apps"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	base := strings.TrimRight(getenv("FORGE_URL", "http://localhost:8080"), "/")
	client := &http.Client{Timeout: 15 * time.Second}

	switch os.Args[1] {
	case "apps":
		if len(os.Args) < 3 || os.Args[2] != "list" {
			usage()
			os.Exit(2)
		}
		if err := listApps(client, base); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "deploy":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: forge deploy <app-name>")
			os.Exit(2)
		}
		if err := deployApp(client, base, os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func listApps(client *http.Client, base string) error {
	var list []apps.App
	if err := getJSON(client, base+"/v1/apps", &list); err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Println("no apps")
		return nil
	}
	fmt.Printf("%-36s  %-16s  %-12s  %s\n", "ID", "NAME", "STATUS", "IMAGE")
	for _, a := range list {
		fmt.Printf("%-36s  %-16s  %-12s  %s\n", a.ID, a.Name, a.Status, a.Image)
	}
	return nil
}

func deployApp(client *http.Client, base, name string) error {
	var list []apps.App
	if err := getJSON(client, base+"/v1/apps", &list); err != nil {
		return err
	}
	var found *apps.App
	for i := range list {
		if list[i].Name == name {
			found = &list[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("app %q not found", name)
	}

	req, err := http.NewRequest(http.MethodPost, base+"/v1/apps/"+found.ID+"/deploy", nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("deploy failed: %s", strings.TrimSpace(string(body)))
	}
	fmt.Printf("deploying %s (%s)\n", found.Name, found.ID)
	return nil
}

func getJSON(client *http.Client, url string, dest any) error {
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("GET %s: %s", url, strings.TrimSpace(string(b)))
	}
	return json.NewDecoder(res.Body).Decode(dest)
}

func usage() {
	fmt.Fprint(os.Stderr, `forge CLI — talks to the Forge API (default http://localhost:8080)

Commands:
  forge apps list
  forge deploy <app-name>

Env:
  FORGE_URL   API base URL (default http://localhost:8080)
`)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
