package agent_test

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/agent"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetRealIPFromRequest(t *testing.T) {
	a := &agent.Agent{}
	req := resty.New().R().SetHeader("X-Real-IP", "192.168.1.1")

	a.SetRealIPFromRequest(req)

	assert.Equal(t, "192.168.1.1", a.RealIP, "Real IP should be set correctly")
}
