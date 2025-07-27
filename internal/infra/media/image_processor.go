package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strings"

	"zmeow/internal/domain/group"
	"zmeow/pkg/logger"

	"github.com/vincent-petithory/dataurl"
)

// ImageProcessor implementa processamento de imagens para grupos
type ImageProcessor struct {
	logger logger.Logger
}

// NewImageProcessor cria uma nova instância do processador de imagens
func NewImageProcessor(logger logger.Logger) *ImageProcessor {
	return &ImageProcessor{
		logger: logger.WithComponent("image-processor"),
	}
}

// ImageFormat representa os formatos de imagem suportados
type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
)

// ImageInfo contém informações sobre uma imagem processada
type ImageInfo struct {
	Format ImageFormat `json:"format"`
	Width  int         `json:"width"`
	Height int         `json:"height"`
	Size   int64       `json:"size"`
	Data   []byte      `json:"-"`
}

// Configurações de processamento
const (
	MaxImageSize   = 5 * 1024 * 1024 // 5MB
	MinImageSize   = 1024            // 1KB (WhatsApp rejeita imagens muito pequenas)
	MaxImageWidth  = 640             // pixels (WhatsApp recomenda max 640)
	MaxImageHeight = 640             // pixels (WhatsApp recomenda max 640)
	MinImageWidth  = 100             // pixels
	MinImageHeight = 100             // pixels
	JPEGQuality    = 90              // qualidade JPEG (0-100) - alta qualidade para compatibilidade
)

// ProcessBase64Image processa uma imagem em formato Base64 (seguindo referência)
func (ip *ImageProcessor) ProcessBase64Image(base64Data string) (*ImageInfo, error) {
	ip.logger.Info().Msg("Processing Base64 image")

	// Validar entrada
	if base64Data == "" {
		return nil, group.NewValidationError("image", "", "image data is required")
	}

	// Verificar se começa com data:image (como na referência)
	if len(base64Data) < 10 || base64Data[0:10] != "data:image" {
		return nil, group.NewValidationError("image", "", "image data should start with \"data:image/\" (supported formats: jpeg, png, gif, webp)")
	}

	// Decodificar usando dataurl (como na referência)
	dataURL, err := dataurl.DecodeString(base64Data)
	if err != nil {
		ip.logger.WithError(err).Error().Msg("Failed to decode base64 data")
		return nil, group.NewValidationError("image", "", "could not decode base64 encoded data from payload")
	}

	filedata := dataURL.Data

	// Validar que temos dados de imagem
	if len(filedata) == 0 {
		return nil, group.NewValidationError("image", "", "no image data found in payload")
	}

	// Validar formato JPEG (WhatsApp requer JPEG) - como na referência
	if len(filedata) < 3 || filedata[0] != 0xFF || filedata[1] != 0xD8 || filedata[2] != 0xFF {
		return nil, group.NewValidationError("image", "", "image must be in JPEG format. WhatsApp only accepts JPEG images for group photos")
	}

	// Validar tamanho do arquivo
	if int64(len(filedata)) > MaxImageSize {
		return nil, group.NewMediaError("image", int64(len(filedata)), "jpeg",
			fmt.Errorf("image size %d exceeds maximum %d bytes", len(filedata), MaxImageSize))
	}

	// Validar tamanho mínimo (WhatsApp rejeita imagens muito pequenas)
	if int64(len(filedata)) < MinImageSize {
		return nil, group.NewValidationError("image", "",
			fmt.Sprintf("image size %d bytes is too small. WhatsApp requires at least %d bytes for group photos", len(filedata), MinImageSize))
	}

	// Retornar dados sem processamento adicional (como na referência)
	imageInfo := &ImageInfo{
		Format: FormatJPEG,
		Width:  0, // Não calculamos dimensões para simplificar
		Height: 0, // Não calculamos dimensões para simplificar
		Size:   int64(len(filedata)),
		Data:   filedata,
	}

	ip.logger.WithFields(map[string]any{
		"format": imageInfo.Format,
		"size":   imageInfo.Size,
	}).Info().Msg("Image processed successfully")

	return imageInfo, nil
}

// isValidFormat verifica se o formato da imagem é suportado
func (ip *ImageProcessor) isValidFormat(format string) bool {
	switch format {
	case "jpeg", "jpg", "png":
		return true
	default:
		return false
	}
}

// ValidateImageData valida dados de imagem brutos
func (ip *ImageProcessor) ValidateImageData(data []byte) error {
	if len(data) == 0 {
		return group.NewValidationError("image", "", "image data is empty")
	}

	if int64(len(data)) > MaxImageSize {
		return group.NewMediaError("image", int64(len(data)), "unknown",
			fmt.Errorf("image size exceeds maximum %d bytes", MaxImageSize))
	}

	// Tentar decodificar para validar formato
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return group.NewMediaError("image", int64(len(data)), "unknown",
			fmt.Errorf("invalid image format: %w", err))
	}

	if !ip.isValidFormat(format) {
		return group.NewMediaError("image", int64(len(data)), format,
			fmt.Errorf("unsupported image format: %s", format))
	}

	return nil
}

// GetImageInfo obtém informações sobre uma imagem sem processá-la
func (ip *ImageProcessor) GetImageInfo(data []byte) (*ImageInfo, error) {
	if err := ip.ValidateImageData(data); err != nil {
		return nil, err
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	return &ImageInfo{
		Format: ImageFormat(format),
		Width:  config.Width,
		Height: config.Height,
		Size:   int64(len(data)),
		Data:   data,
	}, nil
}

// ConvertToJPEG converte uma imagem para formato JPEG
func (ip *ImageProcessor) ConvertToJPEG(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// CreateThumbnail cria uma miniatura da imagem (não implementado nesta versão)
func (ip *ImageProcessor) CreateThumbnail(data []byte, maxWidth, maxHeight int) ([]byte, error) {
	// TODO: Implementar redimensionamento de imagem
	// Por enquanto, retornar a imagem original
	return data, nil
}

// ProcessImageURL baixa e processa uma imagem de uma URL (simplificado)
func (ip *ImageProcessor) ProcessImageURL(imageURL string) (*ImageInfo, error) {
	ip.logger.Info().Str("url", imageURL).Msg("Processing image from URL")

	// Validar URL
	if imageURL == "" {
		return nil, group.NewValidationError("image_url", "", "image URL is required")
	}

	// Fazer download da imagem
	resp, err := http.Get(imageURL)
	if err != nil {
		ip.logger.WithError(err).Error().Msg("Failed to download image from URL")
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status HTTP
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// Verificar Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("URL does not point to an image (Content-Type: %s)", contentType)
	}

	// Ler dados da imagem
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		ip.logger.WithError(err).Error().Msg("Failed to read image data from URL")
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Verificar tamanho
	if len(imageData) > MaxImageSize {
		return nil, group.NewMediaError("image", int64(len(imageData)), "",
			fmt.Errorf("image size %d bytes exceeds maximum %d bytes", len(imageData), MaxImageSize))
	}

	// Validar tamanho mínimo (WhatsApp rejeita imagens muito pequenas)
	if int64(len(imageData)) < MinImageSize {
		return nil, group.NewValidationError("image", "",
			fmt.Sprintf("image size %d bytes is too small. WhatsApp requires at least %d bytes for group photos", len(imageData), MinImageSize))
	}

	// Validar formato JPEG (WhatsApp requer JPEG) - como na referência
	if len(imageData) < 3 || imageData[0] != 0xFF || imageData[1] != 0xD8 || imageData[2] != 0xFF {
		return nil, group.NewValidationError("image", "", "image must be in JPEG format. WhatsApp only accepts JPEG images for group photos")
	}

	// Retornar dados sem processamento adicional (como na referência)
	imageInfo := &ImageInfo{
		Format: FormatJPEG,
		Width:  0, // Não calculamos dimensões para simplificar
		Height: 0, // Não calculamos dimensões para simplificar
		Size:   int64(len(imageData)),
		Data:   imageData,
	}

	ip.logger.WithFields(map[string]any{
		"format": imageInfo.Format,
		"size":   imageInfo.Size,
	}).Info().Msg("Image processed successfully")

	return imageInfo, nil
}
