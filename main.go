package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	clicmd "terracostcli/cmd"
	"text/tabwriter"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type Costing struct {
	Address      string
	ResourceType string
	HourlyCost   float64
	MonthlyCost  float64
}
type Response struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Message  string    `json:"message"`
	Costing  []Costing `json:"costing"`
}

func uploadFile(filename, url, authKey string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set Authorization header from env var
	if authKey != "" {
		req.Header.Set("Authorization", authKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: invalid or missing authorization key")
	}

	var res Response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Resource Name\tResource Type\tHourly Cost\tMonthly Cost\n")
	fmt.Fprintf(w, "-------------\t-------------\t-----------\t------------\n")

	// var costs []Costing

	for _, res := range res.Costing {
		address := res.Address
		resourceType := res.ResourceType
		plannedCostFloatHourly := res.HourlyCost
		plannedCostFloatMonthly := res.MonthlyCost

		fmt.Fprintf(w, "%s\t%s\t%.2f\t%.2f\n",
			address, resourceType, plannedCostFloatHourly, plannedCostFloatMonthly)

		// costs = append(costs, Costing{
		// 	Address:      res.Address,
		// 	ResourceType: res.Type,
		// 	HourlyCost:   plannedCostFloatHourly,
		// 	MonthlyCost:  plannedCostFloatMonthly,
		// })
	}

	w.Flush()
	return nil
}

func main() {
	var (
		authKey string
		url     string
	)
	envFile, _ := godotenv.Read(".env")

	IP := envFile["IP"]
	URL := fmt.Sprintf("http://%s:8080/upload", IP)
	rootCmd := &cobra.Command{
		Use:   "terracost",
		Short: "A CLI tool for getting the cost of your terraform plan",
	}

	uploadCmd := &cobra.Command{
		Use:   "upload [file_path]",
		Short: "Upload the added terraformplan file in json",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			if authKey == "" {
				authKey, _ = clicmd.LoadToken()
			}
			if authKey == "" {
				fmt.Println("No authorization token found. Please run `terracost login` first.")
				os.Exit(1)
			}
			if url == "" {
				url = URL
			}
			err := uploadFile(filename, url, authKey)
			if err != nil {
				fmt.Println("Upload failed:", err)
				os.Exit(1)
			}
		},
	}

	uploadCmd.Flags().StringVarP(&authKey, "auth", "a", "", "Authorization key (or use TerraCost_Authorization_key env var)")
	uploadCmd.Flags().StringVarP(&url, "url", "u", "", "Upload URL")

	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(clicmd.LoginCmd)
	rootCmd.Execute()
}
