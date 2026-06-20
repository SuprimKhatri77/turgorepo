package cloudinary

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Client struct {
	cld *cloudinary.Cloudinary
}

func New(cloudName, apiKey, apiSecret string) (*Client, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary: missing required credentials")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &Client{cld: cld}, nil
}

func (c *Client) UploadImage(ctx context.Context, file interface{}, folder string) (string, string, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:       folder,
		ResourceType: "image",
	})
	if err != nil {
		return "", "", err
	}
	slog.Info("cloudinary response", "full", fmt.Sprintf("%+v", resp))

	return resp.SecureURL, resp.PublicID, nil
}

func (c *Client) DeleteImage(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}
