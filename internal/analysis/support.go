package analysis

import (
	"math"

	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-214: Dynamic support and resistance levels

// SupportResistanceLevel represents a support or resistance level
type SupportResistanceLevel struct {
	Price      float64 `json:"price"`
	Type       string  `json:"type"`       // support, resistance
	Strength   float64 `json:"strength"`   // 0-100
	Touches    int     `json:"touches"`    // number of times price touched this level
	LastTouch  int     `json:"last_touch"` // periods ago
	Confidence float64 `json:"confidence"` // 0-100
}

// SupportResistanceLevels contains all detected levels
type SupportResistanceLevels struct {
	Support    []*SupportResistanceLevel `json:"support"`
	Resistance []*SupportResistanceLevel `json:"resistance"`
	Current    *CurrentLevelInfo         `json:"current"`
}

// CurrentLevelInfo provides context about current price position
type CurrentLevelInfo struct {
	Price                float64                 `json:"price"`
	NearestSupport       *SupportResistanceLevel `json:"nearest_support"`
	NearestResistance    *SupportResistanceLevel `json:"nearest_resistance"`
	DistanceToSupport    float64                 `json:"distance_to_support"`    // percentage
	DistanceToResistance float64                 `json:"distance_to_resistance"` // percentage
	Position             string                  `json:"position"`               // near_support, near_resistance, middle
}

// SupportResistanceDetector detects dynamic support and resistance levels
type SupportResistanceDetector struct {
	tolerance  float64 // price tolerance for level clustering (percentage)
	minTouches int     // minimum touches to consider a level valid
}

// NewSupportResistanceDetector creates a new support/resistance detector
func NewSupportResistanceDetector() *SupportResistanceDetector {
	return &SupportResistanceDetector{
		tolerance:  0.005, // 0.5% tolerance
		minTouches: 2,     // minimum 2 touches
	}
}

// DetectLevels identifies support and resistance levels
func (srd *SupportResistanceDetector) DetectLevels(candles []*models.OHLCV) *SupportResistanceLevels {
	if len(candles) < 10 {
		return &SupportResistanceLevels{}
	}

	// Find pivot points (local highs and lows)
	pivots := srd.findPivotPoints(candles)

	// Cluster nearby levels
	supportClusters := srd.clusterLevels(pivots.lows, candles)
	resistanceClusters := srd.clusterLevels(pivots.highs, candles)

	// Convert clusters to levels
	supportLevels := srd.clustersToLevels(supportClusters, "support", candles)
	resistanceLevels := srd.clustersToLevels(resistanceClusters, "resistance", candles)

	// Filter by minimum touches and sort by strength
	supportLevels = srd.filterAndSortLevels(supportLevels)
	resistanceLevels = srd.filterAndSortLevels(resistanceLevels)

	// Get current price context
	currentPrice := candles[len(candles)-1].Close
	currentInfo := srd.getCurrentLevelInfo(currentPrice, supportLevels, resistanceLevels)

	return &SupportResistanceLevels{
		Support:    supportLevels,
		Resistance: resistanceLevels,
		Current:    currentInfo,
	}
}

// PivotPoints contains pivot high and low points
type PivotPoints struct {
	highs []PivotPoint
	lows  []PivotPoint
}

// PivotPoint represents a pivot high or low
type PivotPoint struct {
	price  float64
	index  int
	candle *models.OHLCV
}

// findPivotPoints identifies pivot highs and lows
func (srd *SupportResistanceDetector) findPivotPoints(candles []*models.OHLCV) *PivotPoints {
	pivots := &PivotPoints{
		highs: make([]PivotPoint, 0),
		lows:  make([]PivotPoint, 0),
	}

	lookback := 3 // periods to look back/forward for pivot confirmation

	for i := lookback; i < len(candles)-lookback; i++ {
		current := candles[i]

		// Check for pivot high
		isPivotHigh := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j != i && candles[j].High >= current.High {
				isPivotHigh = false
				break
			}
		}

		if isPivotHigh {
			pivots.highs = append(pivots.highs, PivotPoint{
				price:  current.High,
				index:  i,
				candle: current,
			})
		}

		// Check for pivot low
		isPivotLow := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j != i && candles[j].Low <= current.Low {
				isPivotLow = false
				break
			}
		}

		if isPivotLow {
			pivots.lows = append(pivots.lows, PivotPoint{
				price:  current.Low,
				index:  i,
				candle: current,
			})
		}
	}

	return pivots
}

// LevelCluster represents a cluster of similar price levels
type LevelCluster struct {
	centerPrice float64
	points      []PivotPoint
	strength    float64
}

// clusterLevels groups nearby price levels together
func (srd *SupportResistanceDetector) clusterLevels(points []PivotPoint, candles []*models.OHLCV) []*LevelCluster {
	if len(points) == 0 {
		return nil
	}

	clusters := make([]*LevelCluster, 0)
	used := make([]bool, len(points))

	for i, point := range points {
		if used[i] {
			continue
		}

		cluster := &LevelCluster{
			centerPrice: point.price,
			points:      []PivotPoint{point},
		}
		used[i] = true

		// Find nearby points within tolerance
		for j := i + 1; j < len(points); j++ {
			if used[j] {
				continue
			}

			priceDiff := math.Abs(points[j].price-point.price) / point.price
			if priceDiff <= srd.tolerance {
				cluster.points = append(cluster.points, points[j])
				used[j] = true

				// Update center price (weighted average)
				totalPrice := 0.0
				for _, p := range cluster.points {
					totalPrice += p.price
				}
				cluster.centerPrice = totalPrice / float64(len(cluster.points))
			}
		}

		// Calculate cluster strength
		cluster.strength = srd.calculateClusterStrength(cluster, candles)

		clusters = append(clusters, cluster)
	}

	return clusters
}

// calculateClusterStrength calculates the strength of a level cluster
func (srd *SupportResistanceDetector) calculateClusterStrength(cluster *LevelCluster, candles []*models.OHLCV) float64 {
	strength := 0.0

	// Base strength from number of touches
	strength += float64(len(cluster.points)) * 20

	// Bonus for recent touches
	currentIndex := len(candles) - 1
	for _, point := range cluster.points {
		age := currentIndex - point.index
		if age < 20 {
			strength += float64(20-age) * 2 // More recent = higher strength
		}
	}

	// Bonus for volume at touch points
	for _, point := range cluster.points {
		avgVolume := srd.calculateAverageVolume(candles, 20)
		if float64(point.candle.Volume) > avgVolume*1.2 {
			strength += 15
		}
	}

	// Cap at 100
	if strength > 100 {
		strength = 100
	}

	return strength
}

// clustersToLevels converts clusters to support/resistance levels
func (srd *SupportResistanceDetector) clustersToLevels(clusters []*LevelCluster, levelType string, candles []*models.OHLCV) []*SupportResistanceLevel {
	levels := make([]*SupportResistanceLevel, 0)

	for _, cluster := range clusters {
		// Find most recent touch
		mostRecentIndex := 0
		for _, point := range cluster.points {
			if point.index > mostRecentIndex {
				mostRecentIndex = point.index
			}
		}

		lastTouch := len(candles) - 1 - mostRecentIndex

		// Calculate confidence based on multiple factors
		confidence := srd.calculateLevelConfidence(cluster, candles)

		level := &SupportResistanceLevel{
			Price:      cluster.centerPrice,
			Type:       levelType,
			Strength:   cluster.strength,
			Touches:    len(cluster.points),
			LastTouch:  lastTouch,
			Confidence: confidence,
		}

		levels = append(levels, level)
	}

	return levels
}

// calculateLevelConfidence calculates confidence in a support/resistance level
func (srd *SupportResistanceDetector) calculateLevelConfidence(cluster *LevelCluster, candles []*models.OHLCV) float64 {
	confidence := 50.0 // Base confidence

	// More touches = higher confidence
	confidence += float64(len(cluster.points)) * 10

	// Recent validation increases confidence
	currentIndex := len(candles) - 1
	for _, point := range cluster.points {
		age := currentIndex - point.index
		if age < 10 {
			confidence += 10
		}
	}

	// Tight clustering increases confidence
	priceVariance := srd.calculatePriceVariance(cluster.points)
	if priceVariance < srd.tolerance*0.5 {
		confidence += 15
	}

	// Cap at 95
	if confidence > 95 {
		confidence = 95
	}

	return confidence
}

// filterAndSortLevels filters levels by minimum touches and sorts by strength
func (srd *SupportResistanceDetector) filterAndSortLevels(levels []*SupportResistanceLevel) []*SupportResistanceLevel {
	// Filter by minimum touches
	filtered := make([]*SupportResistanceLevel, 0)
	for _, level := range levels {
		if level.Touches >= srd.minTouches {
			filtered = append(filtered, level)
		}
	}

	// Sort by strength (descending)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Strength > filtered[i].Strength {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Keep top 5 levels
	if len(filtered) > 5 {
		filtered = filtered[:5]
	}

	return filtered
}

// getCurrentLevelInfo provides context about current price position
func (srd *SupportResistanceDetector) getCurrentLevelInfo(currentPrice float64, supportLevels, resistanceLevels []*SupportResistanceLevel) *CurrentLevelInfo {
	info := &CurrentLevelInfo{
		Price: currentPrice,
	}

	// Find nearest support (below current price)
	var nearestSupport *SupportResistanceLevel
	smallestSupportDistance := math.Inf(1)

	for _, level := range supportLevels {
		if level.Price < currentPrice {
			distance := currentPrice - level.Price
			if distance < smallestSupportDistance {
				smallestSupportDistance = distance
				nearestSupport = level
			}
		}
	}

	// Find nearest resistance (above current price)
	var nearestResistance *SupportResistanceLevel
	smallestResistanceDistance := math.Inf(1)

	for _, level := range resistanceLevels {
		if level.Price > currentPrice {
			distance := level.Price - currentPrice
			if distance < smallestResistanceDistance {
				smallestResistanceDistance = distance
				nearestResistance = level
			}
		}
	}

	info.NearestSupport = nearestSupport
	info.NearestResistance = nearestResistance

	// Calculate distances as percentages
	if nearestSupport != nil {
		info.DistanceToSupport = ((currentPrice - nearestSupport.Price) / currentPrice) * 100
	}

	if nearestResistance != nil {
		info.DistanceToResistance = ((nearestResistance.Price - currentPrice) / currentPrice) * 100
	}

	// Determine position
	if info.DistanceToSupport < 2.0 {
		info.Position = "near_support"
	} else if info.DistanceToResistance < 2.0 {
		info.Position = "near_resistance"
	} else {
		info.Position = "middle"
	}

	return info
}

// Helper methods

func (srd *SupportResistanceDetector) calculateAverageVolume(candles []*models.OHLCV, period int) float64 {
	if len(candles) < period {
		period = len(candles)
	}

	sum := int64(0)
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Volume
	}

	return float64(sum) / float64(period)
}

func (srd *SupportResistanceDetector) calculatePriceVariance(points []PivotPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, point := range points {
		sum += point.price
	}
	mean := sum / float64(len(points))

	// Calculate variance
	variance := 0.0
	for _, point := range points {
		variance += math.Pow(point.price-mean, 2)
	}
	variance /= float64(len(points))

	return math.Sqrt(variance) / mean // Coefficient of variation
}
