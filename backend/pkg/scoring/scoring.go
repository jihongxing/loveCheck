package scoring

import "math"

// Tag weights grouped by risk category.
// Higher weight = more severe risk indicator.
var TagWeights = map[string]float64{
	// Integrity & Identity (1.0–1.8)
	"identity_fraud":  1.5,
	"habitual_lying":  1.2,
	"hidden_marriage": 1.3,
	"child_concealment": 1.2,
	"criminal_record":   1.8,

	// Relational & Moral (0.8–1.3)
	"cheating":        1.0,
	"multi_timing":    1.0,
	"emotional_abuse": 1.2,
	"pua":             1.3,
	"ghosting":        0.8,
	"exploitation":    1.0,

	// Financial & Legal (1.2–2.5)
	"financial_dispute": 1.5,
	"property_fraud":    2.0,
	"romance_scam":      2.5,
	"debt_concealment":  1.5,
	"info_theft":        1.8,
	"gambling":          1.2,
	"drug_abuse":        1.5,

	// Violence & Safety (1.5–3.0)
	"violent_tendency": 3.0,
	"stalking":         2.5,
	"privacy_threat":   2.0,
	"verbal_abuse":     1.5,
	"cyber_bullying":   1.8,
	"over_control":     2.0,

	// Past Life & Background (0.6–1.2)
	"hidden_sex_work":       0.8,
	"hidden_nightlife":      0.6,
	"hidden_disease":        1.0,
	"hidden_marriages":      0.8,
	"hidden_illegal_habits": 1.2,
	"career_fabrication":    0.7,
}

// CategoryNames maps tag keys to their category for breakdown display.
var TagCategory = map[string]string{
	"identity_fraud": "integrity", "habitual_lying": "integrity",
	"hidden_marriage": "integrity", "child_concealment": "integrity",
	"criminal_record": "integrity",

	"cheating": "relational", "multi_timing": "relational",
	"emotional_abuse": "relational", "pua": "relational",
	"ghosting": "relational", "exploitation": "relational",

	"financial_dispute": "financial", "property_fraud": "financial",
	"romance_scam": "financial", "debt_concealment": "financial",
	"info_theft": "financial", "gambling": "financial",
	"drug_abuse": "financial",

	"violent_tendency": "safety", "stalking": "safety",
	"privacy_threat": "safety", "verbal_abuse": "safety",
	"cyber_bullying": "safety", "over_control": "safety",

	"hidden_sex_work": "background", "hidden_nightlife": "background",
	"hidden_disease": "background", "hidden_marriages": "background",
	"hidden_illegal_habits": "background", "career_fabrication": "background",
}

// CategoryMultiplier applies an extra boost for high-severity categories.
var CategoryMultiplier = map[string]float64{
	"integrity":  1.0,
	"relational": 0.9,
	"financial":  1.3,
	"safety":     1.5,
	"background": 0.7,
}

type ScoreBreakdown struct {
	RiskScore      float64            `json:"risk_score"`
	TagScore       float64            `json:"tag_score"`
	ReportBonus    float64            `json:"report_bonus"`
	CategoryScores map[string]float64 `json:"category_scores"`
	RiskLevel      string             `json:"risk_level"`
}

// Calculate computes a weighted R-Score from unique tags and report count.
//
//	tagScore     = sum(tagWeight * categoryMultiplier) for all unique tags
//	reportBonus  = min(3.0, ln(reportCount+1) * 1.2)
//	rawScore     = 1.0 + tagScore*0.6 + reportBonus
//	R-Score      = clamp(rawScore, 1.0, 10.0)
func Calculate(uniqueTags []string, reportCount int) ScoreBreakdown {
	catScores := make(map[string]float64)
	tagScore := 0.0

	for _, tag := range uniqueTags {
		w, ok := TagWeights[tag]
		if !ok {
			w = 1.0
		}
		cat := TagCategory[tag]
		if cat == "" {
			cat = "relational"
		}
		cm := CategoryMultiplier[cat]
		if cm == 0 {
			cm = 1.0
		}
		weighted := w * cm
		tagScore += weighted
		catScores[cat] += weighted
	}

	reportBonus := math.Min(3.0, math.Log(float64(reportCount)+1)*1.2)
	rawScore := 1.0 + tagScore*0.6 + reportBonus
	finalScore := math.Min(10.0, math.Max(1.0, rawScore))
	finalScore = math.Round(finalScore*10) / 10

	level := "low"
	if finalScore >= 8.0 {
		level = "critical"
	} else if finalScore >= 6.0 {
		level = "high"
	} else if finalScore >= 3.5 {
		level = "medium"
	}

	return ScoreBreakdown{
		RiskScore:      finalScore,
		TagScore:       math.Round(tagScore*100) / 100,
		ReportBonus:    math.Round(reportBonus*100) / 100,
		CategoryScores: catScores,
		RiskLevel:      level,
	}
}
