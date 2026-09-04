package export

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestRasterRejectsUnboundedAllocations(t *testing.T) {
	for _, tc := range []struct{ w, h, scale float64 }{{1e20, 10, 1}, {10000, 10000, 1}, {20, 20, math.Inf(1)}, {20, 20, math.NaN()}, {32769, 1, 1}} {
		source := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %g %g"><rect width="1" height="1"/></svg>`, tc.w, tc.h)
		if _, err := RasterizeSVG(source, tc.scale); err == nil {
			t.Fatalf("accepted unsafe dimensions %+v", tc)
		}
	}
	_, err := RasterizeSVG(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10000 10000"/>`, 1)
	if err == nil || !strings.Contains(err.Error(), "reduce the export scale") {
		t.Fatalf("missing actionable limit: %v", err)
	}
}
