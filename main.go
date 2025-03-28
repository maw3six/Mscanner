package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var FR = "\033[31m"   // Red
var FC = "\033[33m"   // Yellow
var FW = "\033[37m"   // White
var FG = "\033[32m"   // Green

var Maw = "mawresult"

var Signs []string
var Strings_Shells []string
var Locations []string
var TrustedFiles []string
var user_agents []string

var headers map[string]string

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return lines, scanner.Err()
}

func banners() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run()

	fmt.Print(`
==== [Shell Finder Advance With Big List Path] ====
`)
	fmt.Println(FC + "[Maw - Scanner] - " + FG + "Nyari Webshell")
}

func URLdomain(site string) string {
	if strings.HasPrefix(site, "http://") {
		site = strings.Replace(site, "http://", "", 1)
	} else if strings.HasPrefix(site, "https://") {
		site = strings.Replace(site, "https://", "", 1)
	}
	pattern := regexp.MustCompile(`(.*)/`)
	for {
		matches := pattern.FindAllStringSubmatch(site, -1)
		if len(matches) == 0 {
			break
		}
		site = matches[0][1]
	}
	return site
}

func IndeXOf(Contents string) bool {
	return strings.Contains(Contents, "<title>Index of")
}

func Send_Request(client *http.Client, url string, Path string) (string, error) {
	req, err := http.NewRequest("GET", url+Path, nil)
	if err != nil {
		return "", err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func Extract_Folders(FoldersName string) bool {
	return !strings.Contains(FoldersName, ".")
}

func Extract_Files(FileName string) bool {
	if strings.Contains(FileName, ".") {
		return strings.Contains(FileName, ".php") || strings.Contains(FileName, ".phtml") || strings.Contains(FileName, ".php5") ||
			strings.Contains(FileName, ".php4") || strings.Contains(FileName, ".phar") || strings.Contains(FileName, ".shtml") ||
			strings.Contains(FileName, ".haxor") || strings.Contains(FileName, ".py") || strings.Contains(FileName, ".env") ||
			strings.Contains(FileName, ".alfa") || strings.Contains(FileName, ".php7")
	}
	return false
}

func Extract(Contents string, Selected string) []string {
	var Pathfiles [][]string
	if strings.Contains(Contents, `</td><td><a href="`) {
		if strings.Contains(Selected, "Files") || strings.Contains(Selected, "Folders") {
			re := regexp.MustCompile(`</td><td><a href="(.*?)">`)
			Pathfiles = re.FindAllStringSubmatch(Contents, -1)
			return extractMatches(Pathfiles)
		}
	} else if strings.Contains(Contents, `]"> <a href="`) {
		if strings.Contains(Selected, "Files") || strings.Contains(Selected, "Folders") {
			re := regexp.MustCompile(`]"> <a href="(.*?)">`)
			Pathfiles = re.FindAllStringSubmatch(Contents, -1)
			return extractMatches(Pathfiles)
		}
	} else if strings.Contains(Contents, "width=device-width, initial-scale=1.0") || strings.Contains(Contents, `<tr><td data-sort=`) {
		if strings.Contains(Selected, "Files") || strings.Contains(Selected, "Folders") {
			re := regexp.MustCompile(`"><a href="(.*?)"><img class="`)
			Pathfiles = re.FindAllStringSubmatch(Contents, -1)
			return extractMatches(Pathfiles)
		}
	}
	return []string{}
}

func extractMatches(matches [][]string) []string {
	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}

func Check_Backdoors(Response string, Sign string) string {
	NullData := ""
	if Response != "" {
		if strings.Contains(Response, Sign) {
			php := "<?php"
			perl := "#!/usr/bin/perl"
			py := "#!/usr/bin/python"
			sh := "#!/bin/bash"
			cgi := "#!/usr/local/bin/perl"
			if !strings.Contains(Response, php) && !strings.Contains(Response, perl) && !strings.Contains(Response, py) &&
				!strings.Contains(Response, sh) && !strings.Contains(Response, cgi) {
				return Sign
			}
		}
	}
	return NullData
}

func Exploiter(client *http.Client, site string, Dirctorys []string, results chan<- string) {
	defer func() {
		_ = recover()
	}()
	url := "https://" + URLdomain(site)
	for _, Path := range Dirctorys {
		contents, err := Send_Request(client, url, Path)
		if err == nil && contents != "" && IndeXOf(contents) {
			ListDirctors := Extract(contents, "Files")
			if ListDirctors != nil {
				for _, elements := range TrustedFiles {
					element := elements + ".php"
					ListDirctors = removeElement(ListDirctors, element)
				}
				for _, MyDir := range ListDirctors {
					if Extract_Files(MyDir) {
						_FirstFilePhP := Path + MyDir
						Request_Text, err := Send_Request(client, url, _FirstFilePhP)
						if err == nil {
							matched := false
							for _, sign := range Signs {
								if Check_Backdoors(Request_Text, sign) != "" {
									for _, Shells := range Strings_Shells {
										if Check_Backdoors(Request_Text, Shells) != "" {
											results <- fmt.Sprintf("[Maw - Scanner] - %s %s [HORE!]", url, FG)
											appendToFile(Maw+"/Shells.txt", url+_FirstFilePhP+"\n")
											sendToTelegram(url + _FirstFilePhP)
											return
										}
									}
									results <- fmt.Sprintf("[Maw - Scanner] - %s %s [HORE!]", url, FG)
									appendToFile(Maw+"/Success.txt", url+_FirstFilePhP+"\n")
									sendToTelegram(url + _FirstFilePhP)
									matched = true
									break
								}
							}
							if matched {
								return
							} else {
								results <- fmt.Sprintf("[Maw - Scanner] - %s %s [Searching ..]", url, FR)
							}
						} else {
							results <- fmt.Sprintf("[Maw - Scanner] - %s %s [Searching ..]", url, FR)
						}
					}
					if Extract_Folders(MyDir) {
						contents2, err := Send_Request(client, url, Path+"/"+MyDir)
						if err == nil {
							ListDirctors2 := Extract(contents2, "Files")
							if ListDirctors2 != nil {
								for _, elements := range TrustedFiles {
									element := elements + ".php"
									ListDirctors2 = removeElement(ListDirctors2, element)
								}
								for _, MyDir2 := range ListDirctors2 {
									if Extract_Files(MyDir2) {
										_NextFilePhP := Path + MyDir2
										Request_Text, err := Send_Request(client, url, _NextFilePhP)
										if err == nil {
											matched := false
											for _, sign := range Signs {
												if Check_Backdoors(Request_Text, sign) != "" {
													for _, Shells := range Strings_Shells {
														if Check_Backdoors(Request_Text, Shells) != "" {
															results <- fmt.Sprintf("[Maw - Scanner] - %s %s [HORE!]", url, FG)
															appendToFile(Maw+"/Shells.txt", url+_NextFilePhP+"\n")
															sendToTelegram(url + _NextFilePhP)
															return
														}
													}
													results <- fmt.Sprintf("[Maw - Scanner] - %s %s [HORE!]", url, FG)
													appendToFile(Maw+"/Success.txt", url+_NextFilePhP+"\n")
													sendToTelegram(url + _NextFilePhP)
													matched = true
													break
												}
											}
											if matched {
												return
											} else {
												results <- fmt.Sprintf("[Maw - Scanner] - %s %s [Searching ..]", url, FR)
											}
										} else {
											results <- fmt.Sprintf("[Maw - Scanner] - %s %s [Searching ..]", url, FR)
										}
									}
								}
							}
						}
					}
				}
			}
		} else {
			results <- fmt.Sprintf("[Maw - Scanner] - %s %s [Searching ..]", url, FR)
		}
	}
}

func removeElement(slice []string, elem string) []string {
	for i, v := range slice {
		if v == elem {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func appendToFile(filename, data string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(data); err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func sendToTelegram(message string) {
	token := "Your Token Telegram"
	chatID := "Chat ID Telegram"
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonPayload)))
	if err != nil {
		fmt.Println("Error sending message to Telegram:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to send message to Telegram. Status code:", resp.StatusCode)
	} else {
		fmt.Println("Message sent to Telegram successfully")
	}
}

func CmsCheckers(site string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}
	Exploiter(client, site, Locations, results)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if _, err := os.Stat(Maw); os.IsNotExist(err) {
		if err := os.Mkdir(Maw, 0755); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	banners()

	fmt.Print("\n" + FR + "[+] " + FC + "IP/DOMAIN LIST: " + FW)
	var filename string
	fmt.Scanln(&filename)
	target, err := readLines(filename)
	if err != nil {
		fmt.Println("[!] File Ga Ada Bang!: " + filename)
		os.Exit(1)
	}

	Signs, err = readLines("core/Shell-Strings.txt")
	if err != nil {
		fmt.Println("Error reading Shell-Strings.txt:", err)
		return
	}
	Strings_Shells, err = readLines("core/Shell-Strings.txt")
	if err != nil {
		fmt.Println("Error reading Shell-Strings.txt:", err)
		return
	}
	Locations, err = readLines("core/Traversals.txt")
	if err != nil {
		fmt.Println("Error reading Path-Locations.txt:", err)
		return
	}
	TrustedFiles, err = readLines("core/Trusted.txt")
	if err != nil {
		fmt.Println("Error reading Trusted-Files.txt:", err)
		return
	}
	uaLines, err := readLines("core/User-Agents.txt")
	if err != nil {
		fmt.Println("Error reading User-Agents.txt:", err)
		return
	}
	for _, line := range uaLines {
		user_agents = append(user_agents, strings.TrimSpace(line))
	}

	var ua string
	if len(user_agents) > 0 {
		ua = user_agents[rand.Intn(len(user_agents))]
	}
	headers = map[string]string{
		"User-Agent":      ua,
		"Content-type":    "*/*",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.5",
		"Connection":      "keep-alive",
	}

	var wg sync.WaitGroup
	results := make(chan string)
	for _, site := range target {
		wg.Add(1)
		go CmsCheckers(site, results, &wg)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}
}
