package synthetictext

import (
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const SchemaVersion = "eventframe.public-fact-text.v1"

type SourceRef struct {
	Publisher string `json:"publisher"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Accessed  string `json:"accessed"`
}

type Oracle struct {
	QueryBeforeCapture string   `json:"query_before_capture"`
	RelevantPriorIDs   []string `json:"relevant_prior_ids"`
	ObsoletePriorIDs   []string `json:"obsolete_prior_ids,omitempty"`
	ExpectedBehavior   string   `json:"expected_behavior"`
}

type Record struct {
	SchemaVersion string                   `json:"schema_version"`
	Split         string                   `json:"split"`
	Topic         string                   `json:"topic"`
	Capture       model.CaptureTurnRequest `json:"capture"`
	Sources       []SourceRef              `json:"sources,omitempty"`
	Oracle        *Oracle                  `json:"oracle,omitempty"`
}

type Manifest struct {
	SchemaVersion        string         `json:"schema_version"`
	Generator            string         `json:"generator"`
	Sessions             int            `json:"sessions"`
	Turns                int            `json:"turns"`
	Queries              int            `json:"queries"`
	Topics               int            `json:"topics"`
	SplitSessions        map[string]int `json:"split_sessions"`
	SplitTurns           map[string]int `json:"split_turns"`
	CorpusSHA256         string         `json:"corpus_sha256"`
	ContainsPrivateText  bool           `json:"contains_private_text"`
	ContainsPII          bool           `json:"contains_pii"`
	ContainsInventedFact bool           `json:"contains_invented_substantive_fact"`
	IdentifierScope      string         `json:"identifier_scope"`
	Derivation           string         `json:"derivation"`
	OracleContract       string         `json:"oracle_contract"`
	Sources              []SourceRef    `json:"sources"`
}

type factPack struct {
	topic          string
	commonClaim    string
	correction     string
	distinctPrompt string
	distinction    string
	query          string
	answer         string
	followup       string
	followupAnswer string
	source         SourceRef
}

var factPacks = []factPack{
	{
		topic:          "Venus and Mercury temperatures",
		commonClaim:    "Mercury must be the hottest planet because it is closest to the Sun.",
		correction:     "NASA identifies Venus as the hottest planet; its dense atmosphere traps heat through a runaway greenhouse effect.",
		distinctPrompt: "Keep orbital position separate from surface temperature: Mercury is the innermost planet, while Venus is the hottest.",
		distinction:    "Mercury is closest to the Sun, but Venus has the highest surface temperature.",
		query:          "Which planet is hottest, and why is the closest planet not the answer?",
		answer:         "Venus is hottest because its atmosphere traps heat; Mercury is closer to the Sun but is not hottest.",
		followup:       "What property distinguishes Mercury in this comparison?",
		followupAnswer: "Mercury is the planet closest to the Sun.",
		source:         source("NASA", "Venus: Facts", "https://science.nasa.gov/venus/venus-facts/"),
	},
	{
		topic:          "weather and climate",
		commonClaim:    "A single unusually cold day proves that a region has a cold climate.",
		correction:     "NOAA distinguishes short-term atmospheric conditions, called weather, from long-period patterns and averages, called climate.",
		distinctPrompt: "Keep a daily observation separate from a climate normal; they use different time scales.",
		distinction:    "Weather describes short-term conditions, while climate summarizes conditions over extended periods.",
		query:          "Does one cold day determine a region's climate?",
		answer:         "No. One cold day is weather; climate is characterized using long-term patterns and averages.",
		followup:       "Then what category does today's temperature belong to?",
		followupAnswer: "Today's temperature is a weather observation.",
		source:         source("NOAA", "What's the Difference Between Weather and Climate?", "https://www.ncei.noaa.gov/news/weather-vs-climate"),
	},
	{
		topic:          "earthquake magnitude and intensity",
		commonClaim:    "An earthquake has one intensity value that is the same everywhere.",
		correction:     "USGS explains that an earthquake has one magnitude, while shaking intensity varies by location.",
		distinctPrompt: "Keep source size separate from observed local shaking.",
		distinction:    "Magnitude describes earthquake size at the source; intensity describes shaking at a particular place.",
		query:          "Which earthquake measure varies from place to place?",
		answer:         "Intensity varies by location; magnitude is the event's source-size measure.",
		followup:       "What does magnitude represent in that distinction?",
		followupAnswer: "Magnitude represents the size of the earthquake at its source.",
		source:         source("USGS", "What is the difference between earthquake magnitude and earthquake intensity?", "https://www.usgs.gov/faqs/what-difference-between-earthquake-magnitude-and-earthquake-intensity-what-modified-mercalli"),
	},
	{
		topic:          "mass and weight",
		commonClaim:    "Mass and weight are the same physical quantity and both use kilograms in SI.",
		correction:     "NIST distinguishes mass from weight: mass is measured in kilograms, while scientific weight is a force measured in newtons.",
		distinctPrompt: "Keep inertial mass separate from the force associated with gravity.",
		distinction:    "The kilogram is the SI unit of mass; the newton is the SI unit of weight treated as force.",
		query:          "Which SI unit belongs to scientific weight rather than mass?",
		answer:         "The newton belongs to weight as force; the kilogram belongs to mass.",
		followup:       "Which quantity remains measured in kilograms?",
		followupAnswer: "Mass is measured in kilograms.",
		source:         source("NIST", "SI Units - Mass", "https://www.nist.gov/pml/owm/si-units-mass"),
	},
	{
		topic:          "Arctic and Antarctic geography",
		commonClaim:    "The Arctic and Antarctic have the same land-and-ocean arrangement.",
		correction:     "NOAA describes opposite arrangements: the Arctic is an ocean surrounded by continents, while Antarctica is a continent surrounded by oceans.",
		distinctPrompt: "Keep the northern ocean-centered region separate from the southern continent-centered region.",
		distinction:    "The Arctic centers on an ocean; Antarctica is a continent.",
		query:          "Which polar region is a continent surrounded by oceans?",
		answer:         "Antarctica is the continent surrounded by oceans; the Arctic is an ocean surrounded by continents.",
		followup:       "What is at the center of the Arctic arrangement?",
		followupAnswer: "An ocean is at the center of the Arctic arrangement.",
		source:         source("NOAA", "Polar Opposites: the Arctic and Antarctic", "https://prod-01-asg-www-climate.woc.noaa.gov/news-features/understanding-climate/polar-opposites-arctic-and-antarctic"),
	},
	{
		topic:          "meteoroids, meteors, and meteorites",
		commonClaim:    "A streak of light in the atmosphere is a meteorite.",
		correction:     "NASA distinguishes the terms: a meteor is the atmospheric light phenomenon, while a meteorite is material that survives to the ground.",
		distinctPrompt: "Keep the object in space, the atmospheric phenomenon, and surviving ground material as separate stages.",
		distinction:    "A meteoroid is in space, a meteor is observed in an atmosphere, and a meteorite reaches the surface.",
		query:          "What is the correct term for the visible streak in the atmosphere?",
		answer:         "The visible atmospheric streak is a meteor, not a meteorite.",
		followup:       "What do we call material that survives to the ground?",
		followupAnswer: "Material that survives to the ground is called a meteorite.",
		source:         source("NASA JPL", "Asteroid Watch: Fast Facts", "https://www.jpl.nasa.gov/asteroid-watch/fast-facts/"),
	},
	{
		topic:          "cause of Earth's seasons",
		commonClaim:    "Earth's seasons are caused mainly by its changing distance from the Sun.",
		correction:     "NASA explains that Earth's axial tilt causes the seasons; distance from the Sun is not their main cause.",
		distinctPrompt: "Keep orbital distance variation separate from the changing angle and duration of sunlight caused by axial tilt.",
		distinction:    "Earth's tilt drives the seasonal cycle, while its modest distance variation does not explain opposite hemispheric seasons.",
		query:          "What causes Earth's seasons according to NASA?",
		answer:         "Earth's tilted axis causes the seasons by changing sunlight angle and duration through the year.",
		followup:       "Is changing Earth-Sun distance the main seasonal mechanism?",
		followupAnswer: "No. NASA identifies axial tilt, not changing distance, as the main mechanism.",
		source:         source("NASA", "What Causes the Seasons?", "https://spaceplace.nasa.gov/seasons/en/"),
	},
	{
		topic:          "Great Wall visibility from space",
		commonClaim:    "The Great Wall of China is plainly visible to the unaided eye from the Moon.",
		correction:     "NASA states that the Great Wall is not visible from the Moon and is difficult or impossible to see unaided from Earth orbit.",
		distinctPrompt: "Keep images made with high-powered lenses separate from unaided human visibility.",
		distinction:    "A photographed feature under magnification does not establish unaided visibility from the Moon.",
		query:          "Can an unaided observer see the Great Wall from the Moon?",
		answer:         "No. NASA says the Great Wall is not visible from the Moon with unaided vision.",
		followup:       "What distinction matters when orbit photographs show the wall?",
		followupAnswer: "The distinction is between aided photography and unaided human visibility.",
		source:         source("NASA", "Great Wall", "https://www.nasa.gov/image-article/great-wall/"),
	},
	{
		topic:          "bat vision",
		commonClaim:    "All bats are blind and rely only on echolocation.",
		correction:     "The Smithsonian states that bats are not blind; all bats can see, although visual ability varies among species.",
		distinctPrompt: "Keep vision and echolocation as separate sensory capabilities rather than treating one as proof that the other is absent.",
		distinction:    "Bats can see, and many species also use echolocation.",
		query:          "Are bats blind according to the Smithsonian?",
		answer:         "No. Bats can see, though eyesight and reliance on other senses vary by species.",
		followup:       "Does echolocation imply that bats have no vision?",
		followupAnswer: "No. Echolocation does not imply an absence of vision.",
		source:         source("Smithsonian Institution", "Bat Facts", "https://www.si.edu/spotlight/bats/batfacts"),
	},
	{
		topic:          "repeated lightning strikes",
		commonClaim:    "Lightning never strikes the same place twice.",
		correction:     "NOAA's National Severe Storms Laboratory says lightning can strike the same or nearly the same place more than once.",
		distinctPrompt: "Keep a memorable saying separate from observed lightning behavior and safety guidance.",
		distinction:    "Repeated strikes are possible, especially where site characteristics make strikes more likely.",
		query:          "Can lightning strike the same place more than once?",
		answer:         "Yes. NOAA states that repeated strikes at the same or nearly the same place can occur.",
		followup:       "Should the folk saying be used as a safety assumption?",
		followupAnswer: "No. The saying is contradicted by observed repeated strikes.",
		source:         source("NOAA NSSL", "Severe Weather 101: Lightning FAQ", "https://www.nssl.noaa.gov/education/svrwx101/lightning/faq/"),
	},
}

// Build creates generated conversations around cited public facts. It reads no
// private replay data and invents no substantive factual claim.
func Build() ([]Record, Manifest) {
	const sessions = 32
	records := make([]Record, 0, sessions*12)
	splitSessions := map[string]int{"design": 0, "confirmation": 0}
	splitTurns := map[string]int{"design": 0, "confirmation": 0}
	queries := 0

	for sessionIndex := 0; sessionIndex < sessions; sessionIndex++ {
		split := "design"
		if sessionIndex >= 24 {
			split = "confirmation"
		}
		splitSessions[split]++
		pack := factPacks[sessionIndex%len(factPacks)]
		other := factPacks[(sessionIndex+3)%len(factPacks)]
		sessionID := fmt.Sprintf("public-fact-%s-%03d", split, sessionIndex+1)
		start := time.Date(2026, 2, 2+sessionIndex, 10, 0, 0, 0, time.UTC)
		ids := make([]string, 0, 12)

		appendTurn := func(scenario, user, assistant string, sources []SourceRef, oracle *Oracle) {
			turnNumber := len(ids) + 1
			id := fmt.Sprintf("%s-t%02d", sessionID, turnNumber)
			ids = append(ids, id)
			occurred := start.Add(time.Duration(turnNumber-1) * 8 * time.Hour)
			if oracle != nil {
				queries++
			}
			records = append(records, Record{
				SchemaVersion: SchemaVersion, Split: split, Topic: pack.topic, Sources: sources, Oracle: oracle,
				Capture: model.CaptureTurnRequest{
					ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id,
					Turn: model.TurnCapture{
						ID: id, TenantID: "public-fact-corpus", SessionID: sessionID, Sequence: uint64(turnNumber),
						RunID: "public-fact-run-" + sessionID, AgentID: "evaluation-agent",
						UserText: user, AssistantText: assistant,
						OccurredAt: occurred, ObservedAt: occurred.Add(2 * time.Second), AvailableAt: occurred.Add(3 * time.Second),
					},
				},
			})
			splitTurns[split]++
		}

		appendTurn("source_request", "Create a short study note about "+pack.topic+" using the cited public source.", pack.correction, []SourceRef{pack.source}, nil)
		appendTurn("ordinary_public_fact", "Add an unrelated public-science note so retrieval has a realistic distractor.", other.distinction, []SourceRef{other.source}, nil)
		appendTurn("misconception_record", "A draft contains this common claim: "+pack.commonClaim+" Preserve it as a claim awaiting verification.", "Recorded as an unverified claim, not as accepted knowledge.", []SourceRef{pack.source}, nil)
		appendTurn("source_correction", "Verify that draft claim against the cited source and state the correction.", pack.correction, []SourceRef{pack.source}, nil)
		appendTurn("corrected_recall_query", pack.query, pack.answer, []SourceRef{pack.source}, &Oracle{QueryBeforeCapture: pack.query, RelevantPriorIDs: []string{ids[0], ids[3]}, ObsoletePriorIDs: []string{ids[2]}, ExpectedBehavior: "promote_sourced_correction_and_demote_misconception"})
		appendTurn("related_distinction", "Record the nearby distinction without merging the two concepts.", pack.distinctPrompt+" "+pack.distinction, []SourceRef{pack.source}, nil)
		appendTurn("anti_pigeon_query", pack.followup, pack.followupAnswer, []SourceRef{pack.source}, &Oracle{QueryBeforeCapture: pack.followup, RelevantPriorIDs: []string{ids[5]}, ObsoletePriorIDs: []string{ids[2]}, ExpectedBehavior: "separate_related_but_noninterchangeable_concepts"})
		appendTurn("ordinary_public_fact", "Give me one more unrelated fact from the other cited topic.", other.correction, []SourceRef{other.source}, nil)
		appendTurn("delayed_source_confirmation", "I checked the cited page later. Treat the sourced correction as confirmed and keep the original draft claim marked incorrect.", "The correction remains the accepted answer, and the earlier misconception remains excluded from factual recall.", []SourceRef{pack.source}, nil)
		confirmationQuery := "After checking the cited source, what conclusion should replace the earlier draft claim about " + pack.topic + "?"
		appendTurn("delayed_confirmation_query", confirmationQuery, pack.answer, []SourceRef{pack.source}, &Oracle{QueryBeforeCapture: confirmationQuery, RelevantPriorIDs: []string{ids[0], ids[3], ids[4], ids[8]}, ObsoletePriorIDs: []string{ids[2]}, ExpectedBehavior: "join_paraphrased_confirmation_to_sourced_correction"})
		appendTurn("lexical_distractor", "Summarize why labels that sound similar should still be checked against their definitions.", "Similar wording does not make two concepts interchangeable; the source-defined distinction controls.", nil, nil)
		appendTurn("narrow_followup_query", "What was the corrected answer again?", pack.answer, []SourceRef{pack.source}, &Oracle{QueryBeforeCapture: "What was the corrected answer again?", RelevantPriorIDs: []string{ids[3], ids[4], ids[8], ids[9]}, ObsoletePriorIDs: []string{ids[2]}, ExpectedBehavior: "resolve_narrow_followup_from_session_context"})
	}

	sources := make([]SourceRef, 0, len(factPacks))
	for _, pack := range factPacks {
		sources = append(sources, pack.source)
	}
	return records, Manifest{
		SchemaVersion: SchemaVersion, Generator: "go run ./cmd/eventframe-synthetic-text",
		Sessions: sessions, Turns: len(records), Queries: queries, Topics: len(factPacks),
		SplitSessions: splitSessions, SplitTurns: splitTurns,
		ContainsPrivateText: false, ContainsPII: false, ContainsInventedFact: false,
		IdentifierScope: "Deterministic dataset-local labels only. No identifier maps to a person, account, host, production session, or external system.",
		Derivation:      "Newly authored question-and-answer turns grounded in cited public facts from NASA, NOAA, USGS, NIST, and Smithsonian pages. Private replay data informed only the mix of retrieval situations; no private wording or identifiers were read or copied.",
		OracleContract:  "Oracle and source metadata are evaluation-only. Send only the nested capture object to eventframed; never index oracle fields or source annotations as conversation memory.",
		Sources:         sources,
	}
}

func source(publisher, title, url string) SourceRef {
	return SourceRef{Publisher: publisher, Title: title, URL: url, Accessed: "2026-08-30"}
}
