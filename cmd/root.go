package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"

	"github.com/spf13/cobra"
)

const (
	appDesc = "geminipro"
)

var (
	model  string
	apikey string
	port   int
)

func init() {
	initFlags()
}

const (
	modelEnv  = "MODEL"
	apikeyEnv = "APIKEY"
	portEnv   = "PORT"
)

func initFlags() {
	// gemini-2.0-flash-thinking-exp
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", envOrDefaultString(modelEnv, "gemini-2.0-flash-exp"), "model")
	rootCmd.PersistentFlags().StringVarP(&apikey, "apikey", "k", envOrDefaultString(apikeyEnv, ""), "key")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", envOrDefaultInt(portEnv, 3000), "port of server")
}

var rootCmd = &cobra.Command{
	Use:   "geminipro",
	Short: appDesc,
	Long:  appDesc + "long",
	Run: func(cmd *cobra.Command, args []string) {
		if apikey == "" {
			fmt.Printf("%v\n", "key is empty")
			os.Exit(1)
		}
		ctx := context.Background()

		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apikey,
			Backend: genai.BackendGoogleAI,
		})
		if err != nil {
			return
		}
		s := gin.Default()

		s.POST("/api/generate", func(c *gin.Context) {
			req := struct {
				Message []*genai.Content `json:"messages"`
			}{}
			err := c.ShouldBindJSON(&req)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}

			c.Stream(func(w io.Writer) bool {
				for v, err := range client.Models.GenerateContentStream(c, model, req.Message, nil) {
					if err == nil {
						for _, ca := range v.Candidates {
							for _, cc := range ca.Content.Parts {
								// fmt.Printf("%v\n", cc.Text)
								_, err := w.Write([]byte(cc.Text))
								c.Writer.Flush()
								if err != nil {
									return false
								}
							}
						}
						if v.UsageMetadata.TotalTokenCount > 0 {
							return false
						}
					} else {
						w.Write([]byte(err.Error()))
						c.Writer.Flush()
						return false
					}
				}
				return false
				// return true
			})
		})
		s.Run(fmt.Sprintf(":%d", port))
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func envOrDefaultString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrDefaultInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		b, _ := strconv.ParseInt(v, 10, 64)
		return int(b)
	}
	return def
}
