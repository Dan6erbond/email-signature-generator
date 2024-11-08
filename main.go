package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Dan6erbond/email-signature-generator/model"
	"github.com/Dan6erbond/email-signature-generator/views/pages"
	"github.com/Dan6erbond/email-signature-generator/views/pages/components"
	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

func main() {
	k.Load(confmap.Provider(map[string]any{
		"server": map[string]any{
			"port": 8000,
		},
	}, ""), nil)

	k.Load(file.Provider("config.json"), json.Parser())
	k.Load(file.Provider("config.yml"), yaml.Parser())

	k.Load(env.Provider("EMAIL_SIGNATURE_GENERATOR_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "EMAIL_SIGNATURE_GENERATOR_")), "_", ".", -1)
	}), nil)

	minioClient, err := minio.New(k.String("s3.host"), &minio.Options{
		Creds:  credentials.NewStaticV4(k.String("s3.accessKeyId"), k.String("s3.secretAccessKey"), ""),
		Secure: false,
	})

	if err != nil {
		log.Fatalln(err)
	}

	e := echo.New()

	e.Static("/public", "public")

	e.GET("/", func(c echo.Context) error {
		buf := templ.GetBuffer()
		defer templ.ReleaseBuffer(buf)

		home := pages.Home()
		if err := home.Render(c.Request().Context(), buf); err != nil {
			return err
		}

		return c.HTML(http.StatusOK, buf.String())
	})

	e.POST("/", func(c echo.Context) error {
		buf := templ.GetBuffer()
		defer templ.ReleaseBuffer(buf)

		signatureModel := model.Signature{
			Name:        c.FormValue("name"),
			Role:        c.FormValue("role"),
			Email:       c.FormValue("email"),
			PhoneNumber: c.FormValue("phone"),
			LinkedInURL: c.FormValue("linkedin"),
			Company: model.Company{
				Name: k.String("signature.company.name"),
				URL:  k.String("signature.company.url"),
				Address: model.Address{
					Street: k.String("signature.company.address.street"),
					Number: k.String("signature.company.address.number"),
					Zip:    k.String("signature.company.address.zip"),
					Area:   k.String("signature.company.address.area"),
				},
			},
			BrandColor: k.String("signature.brandColor"),
		}

		picture, err := c.FormFile("picture")
		if err != nil && err.Error() != "http: no such file" {
			slog.Error("Error reading picture", slog.Attr{Key: "error", Value: slog.AnyValue(err)})
			return err
		}

		var pictureURL *url.URL

		if picture != nil {
			src, err := picture.Open()
			if err != nil {
				slog.Error("Error opening picture")
				return err
			}

			contentType := "application/octet-stream"

			id := uuid.New()

			// Upload the test file with FPutObject
			info, err := minioClient.PutObject(c.Request().Context(), k.String("s3.bucket"), id.String(), src, -1, minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				slog.Error("Error uploading picture", slog.Attr{Key: "error", Value: slog.AnyValue(err)})
			}

			// Set request parameters
			reqParams := make(url.Values)
			reqParams.Set("response-content-disposition", "attachment; filename=\"your-filename.txt\"")

			pictureURL = &url.URL{
				Host: "localhost:9005",
			}
			pictureURL.Scheme = "http"
			pictureURL.Path = fmt.Sprintf("/%s/%s", k.String("s3.bucket"), info.Key)

			signatureModel.Picture = pictureURL.String()
		}

		if err := components.Signature(signatureModel).Render(c.Request().Context(), buf); err != nil {
			return err
		}

		return c.HTML(http.StatusOK, buf.String())
	})

	err = e.Start(fmt.Sprintf(":%d", k.Int("server.port")))

	if err != nil {
		log.Fatalln("Error launching server", err)
	}
}
