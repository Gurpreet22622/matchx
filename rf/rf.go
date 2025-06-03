package rf

import (
	"log"
	"matchx/models"
	"net/http"
	"strconv"
	"strings"

	ort "github.com/yalue/onnxruntime_go"
)

func mapFurnishingStatus(status string) float32 {
	switch strings.ToLower(status) {
	case "furnished":
		return 2
	case "semi-furnished":
		return 1
	case "unfurnished":
		return 0
	default:
		return -1
	}
}

func mapLeaseType(lease string) float32 {
	switch strings.ToLower(lease) {
	case "bachelors":
		return 0
	case "family":
		return 1
	case "both":
		return 2
	default:
		return -1
	}
}

func PredictRent(
	propertyType, furnishingStatus, leaseType string,
	internet, ac, ro, kitchen, geyser bool,
	localityEncoded, latitude, longitude, propertyArea float32,
) (float32, *models.ResponseError) {

	boolToFloatMap := map[bool]float32{true: 1.0, false: 0.0}
	lastDigit, _ := strconv.ParseFloat(string(propertyType[len(propertyType)-1]), 32)
	pT := float32(lastDigit)
	features := []float32{
		pT, mapFurnishingStatus(furnishingStatus), mapLeaseType(leaseType), propertyArea,
		boolToFloatMap[internet], boolToFloatMap[ac], boolToFloatMap[ro], boolToFloatMap[kitchen], boolToFloatMap[geyser],
		localityEncoded, latitude, longitude,
	}

	ort.SetSharedLibraryPath("/home/gurpreet/onnxruntime-linux-x64-1.21.0/lib/libonnxruntime.so")
	if err := ort.InitializeEnvironment(); err != nil {
		log.Println("ONNX env init error:", err)
		return 0, &models.ResponseError{
			Message: "Failed to initialize ONNX runtime",
			Status:  http.StatusInternalServerError,
		}
	}
	defer ort.DestroyEnvironment()

	inputTensor, err := ort.NewTensor(ort.NewShape(1, 12), features)
	if err != nil {
		log.Println("Input tensor error:", err)
		return 0, &models.ResponseError{
			Message: "Invalid input tensor",
			Status:  http.StatusInternalServerError,
		}
	}
	defer inputTensor.Destroy()

	outputTensor, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 1))
	if err != nil {
		log.Println("Output tensor error:", err)
		return 0, &models.ResponseError{
			Message: "Failed to create output tensor",
			Status:  http.StatusInternalServerError,
		}
	}
	defer outputTensor.Destroy()

	session, err := ort.NewAdvancedSession(
		"rf/random_forest_model.onnx",
		[]string{"float_input"},
		[]string{"variable"},
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		log.Println("Model session error:", err)
		return 0, &models.ResponseError{
			Message: "Failed to create model session",
			Status:  http.StatusInternalServerError,
		}
	}
	defer session.Destroy()

	// Run inference
	if err := session.Run(); err != nil {
		log.Println("Model run error:", err)
		return 0, &models.ResponseError{
			Message: "Failed to run prediction",
			Status:  http.StatusInternalServerError,
		}
	}

	predictedRent := outputTensor.GetData()[0]
	return predictedRent, nil
}
