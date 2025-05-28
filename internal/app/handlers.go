package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kegliz/qplay/qc/builder"
	"github.com/kegliz/qplay/qc/circuit"
	"github.com/kegliz/qplay/qc/renderer"
	"github.com/kegliz/qplay/qc/simulator"

	// Import simulators to register them
	_ "github.com/kegliz/qplay/qc/simulator/itsu"
	_ "github.com/kegliz/qplay/qc/simulator/qsim"
)

// CircuitRequest represents the structure for circuit execution requests
type CircuitRequest struct {
	Circuit struct {
		Qubits int `json:"qubits"`
		Gates  []struct {
			Type   string `json:"type"`
			Qubits []int  `json:"qubits"`
			Step   int    `json:"step"`
		} `json:"gates"`
	} `json:"circuit"`
	Backend string `json:"backend"`
	Shots   int    `json:"shots"`
}

// CircuitResponse represents the structure for circuit execution responses
type CircuitResponse struct {
	Measurements  map[string]int    `json:"measurements,omitempty"`
	StateVector   []complex128      `json:"state_vector,omitempty"`
	CircuitImage  string           `json:"circuit_image,omitempty"`
	ExecutionTime float64          `json:"execution_time,omitempty"`
	Backend       string           `json:"backend"`
	Shots         int              `json:"shots"`
}

var badRequestErrorMsg = "Bad Request - please contact the administrator"
var internalServerErrorMsg = "Internal Server Error - please contact the administrator"

// RootHandler is the handler for the / endpoint
func (a *appServer) RootHandler(c *gin.Context) {
	l, err := a.getLoggerFromContext(c)
	if err != nil {
		panic("logger not found in context")
	}
	l.Debug().Msg("serving root endpoint")

	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "Quantum Playground DEV"})
}

// HealthHandler is the handler for the /health endpoint
func (a *appServer) HealthHandler(c *gin.Context) {
	l, err := a.getLoggerFromContext(c)
	if err != nil {
		panic("logger not found in context")
	}
	l.Debug().Msg("serving health endpoint")
	c.String(http.StatusOK, "OK")
}

// ExecuteCircuit is the handler for the /api/execute endpoint
func (a *appServer) ExecuteCircuit(c *gin.Context) {
	l, err := a.getLoggerFromContext(c)
	if err != nil {
		panic("logger not found in context")
	}
	l.Debug().Msg("serving circuit execution endpoint")

	var req CircuitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error().Err(err).Msg("binding JSON failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if req.Circuit.Qubits <= 0 || req.Circuit.Qubits > 10 {
		l.Error().Int("qubits", req.Circuit.Qubits).Msg("invalid qubit count")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid qubit count (1-10 allowed)"})
		return
	}

	if req.Shots <= 0 || req.Shots > 10000 {
		req.Shots = 1000 // Default value
	}

	if req.Backend == "" {
		req.Backend = "qsim" // Default backend
	}

	// Build circuit from request
	circ, err := a.buildCircuitFromRequest(&req)
	if err != nil {
		l.Error().Err(err).Msg("building circuit failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to build circuit: " + err.Error()})
		return
	}

	// Execute circuit
	result, err := a.executeCircuit(circ, req.Backend, req.Shots)
	if err != nil {
		l.Error().Err(err).Str("backend", req.Backend).Msg("circuit execution failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Circuit execution failed: " + err.Error()})
		return
	}

	// Generate circuit image
	circuitImage, err := a.generateCircuitImage(circ)
	if err != nil {
		l.Warn().Err(err).Msg("failed to generate circuit image")
		// Continue without image - not critical
	}

	// Prepare response
	response := CircuitResponse{
		Measurements:  result,
		CircuitImage:  circuitImage,
		Backend:       req.Backend,
		Shots:         req.Shots,
	}

	c.JSON(http.StatusOK, response)
}

// buildCircuitFromRequest converts the JSON request into a quantum circuit
func (a *appServer) buildCircuitFromRequest(req *CircuitRequest) (circuit.Circuit, error) {
	// Create builder with specified qubits and classical bits
	b := builder.New(builder.Q(req.Circuit.Qubits), builder.C(req.Circuit.Qubits))

	// Sort gates by step to ensure proper order
	gatesByStep := make(map[int][]struct {
		Type   string `json:"type"`
		Qubits []int  `json:"qubits"`
		Step   int    `json:"step"`
	})

	for _, gate := range req.Circuit.Gates {
		gatesByStep[gate.Step] = append(gatesByStep[gate.Step], gate)
	}

	// Add gates in order
	for step := 0; step < 10; step++ {
		gates := gatesByStep[step]
		for _, gate := range gates {
			switch gate.Type {
			case "H":
				if len(gate.Qubits) != 1 {
					return nil, fmt.Errorf("H gate requires exactly 1 qubit")
				}
				b.H(gate.Qubits[0])
			case "X":
				if len(gate.Qubits) != 1 {
					return nil, fmt.Errorf("X gate requires exactly 1 qubit")
				}
				b.X(gate.Qubits[0])
			case "Y":
				if len(gate.Qubits) != 1 {
					return nil, fmt.Errorf("Y gate requires exactly 1 qubit")
				}
				b.Y(gate.Qubits[0])
			case "Z":
				if len(gate.Qubits) != 1 {
					return nil, fmt.Errorf("Z gate requires exactly 1 qubit")
				}
				b.Z(gate.Qubits[0])
			case "S":
				if len(gate.Qubits) != 1 {
					return nil, fmt.Errorf("S gate requires exactly 1 qubit")
				}
				b.S(gate.Qubits[0])
			case "CNOT":
				if len(gate.Qubits) != 2 {
					return nil, fmt.Errorf("CNOT gate requires exactly 2 qubits")
				}
				b.CNOT(gate.Qubits[0], gate.Qubits[1])
			case "CZ":
				if len(gate.Qubits) != 2 {
					return nil, fmt.Errorf("CZ gate requires exactly 2 qubits")
				}
				b.CZ(gate.Qubits[0], gate.Qubits[1])
			case "SWAP":
				if len(gate.Qubits) != 2 {
					return nil, fmt.Errorf("SWAP gate requires exactly 2 qubits")
				}
				b.SWAP(gate.Qubits[0], gate.Qubits[1])
			case "MEASURE":
				if len(gate.Qubits) != 1 {
					return nil, fmt.Errorf("MEASURE requires exactly 1 qubit")
				}
				b.Measure(gate.Qubits[0], gate.Qubits[0])
			default:
				return nil, fmt.Errorf("unsupported gate type: %s", gate.Type)
			}
		}
	}

	// Automatically add measurements if none specified
	hasMeasurements := false
	for _, gate := range req.Circuit.Gates {
		if gate.Type == "MEASURE" {
			hasMeasurements = true
			break
		}
	}

	if !hasMeasurements {
		for i := 0; i < req.Circuit.Qubits; i++ {
			b.Measure(i, i)
		}
	}

	return b.BuildCircuit()
}

// executeCircuit runs the circuit on the specified backend
func (a *appServer) executeCircuit(circ circuit.Circuit, backend string, shots int) (map[string]int, error) {
	// Create runner for the specified backend
	runner, err := simulator.CreateRunner(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s runner: %w", backend, err)
	}

	// Create simulator
	sim := simulator.NewSimulator(simulator.SimulatorOptions{
		Shots:  shots,
		Runner: runner,
	})

	// Run simulation
	results, err := sim.RunSerial(circ)
	if err != nil {
		return nil, fmt.Errorf("simulation failed: %w", err)
	}

	return results, nil
}

// generateCircuitImage creates a PNG image of the circuit
func (a *appServer) generateCircuitImage(circ circuit.Circuit) (string, error) {
	// Create renderer
	r := renderer.NewRenderer(60) // 60 DPI for web display

	// Render circuit to image
	img, err := r.Render(circ)
	if err != nil {
		return "", fmt.Errorf("failed to render circuit: %w", err)
	}

	// Create a buffer to capture the PNG
	var buf bytes.Buffer

	// Encode image as PNG to buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded, nil
}

// CreateCircuit is the handler for the /api/qprogs endpoint
func (a *appServer) CreateCircuit(c *gin.Context) {
	l, err := a.getLoggerFromContext(c)
	if err != nil {
		panic("logger not found in context")
	}
	l.Debug().Msg("serving qprog creation endpoint")
	// var params qservice.ProgramValue
	// if err := c.ShouldBindJSON(&params); err != nil {
	// 	l.Error().Err(err).Msg("binding json failed")
	// 	c.String(http.StatusBadRequest, badRequestErrorMsg)
	// 	return
	// }
	// // Save the circuit
	// id, err := a.qs.SaveProgram(l, &params)
	// if err != nil {
	// 	l.Error().Err(err).Msg("saving circuit failed")
	// 	c.String(http.StatusInternalServerError, internalServerErrorMsg)
	// 	return
	// }
	// c.PureJSON(http.StatusOK, qservice.ProgramIDValue{ID: id})
}

// RenderCircuit is the handler for the /api/qprogs/:id/img endpoint
func (a *appServer) RenderCircuit(c *gin.Context) {
	l, err := a.getLoggerFromContext(c)
	if err != nil {
		panic("logger not found in context")
	}
	l.Debug().Msg("serving rendering circuit img endpoint")
	// id := c.Param("id")
	// img, err := a.qs.RenderCircuit(l, id)
	// if err != nil {
	// 	l.Error().Err(err).Msg("rendering circuit failed")
	// 	c.String(http.StatusInternalServerError, internalServerErrorMsg)
	// 	return
	// }
	// c.Header("Content-Type", "image/png")
	// png.Encode(c.Writer, img)
	// c.Status(http.StatusOK)
}
