-- name: RecordSignal :exec
INSERT INTO personalization_signals (
    id, user_id, trip_id, signal_type, subject_type, subject_id, strength, source_event, occurred_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT DO NOTHING;

-- name: ListSignals :many
SELECT id, user_id, trip_id, signal_type, subject_type, subject_id, strength, source_event, occurred_at
FROM personalization_signals
WHERE user_id = $1 AND occurred_at >= $2
ORDER BY occurred_at DESC
LIMIT $3;

-- name: DeleteSignals :exec
DELETE FROM personalization_signals WHERE user_id = $1;

-- name: UpsertLearnedPreference :exec
INSERT INTO learned_preferences (
    id, user_id, preference_key, confidence, score, evidence_count,
    explanation_code, status, rules_version, first_observed_at, last_observed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (user_id, preference_key) DO UPDATE SET
    confidence = EXCLUDED.confidence,
    score = EXCLUDED.score,
    evidence_count = EXCLUDED.evidence_count,
    explanation_code = EXCLUDED.explanation_code,
    last_observed_at = EXCLUDED.last_observed_at,
    rules_version = EXCLUDED.rules_version,
    updated_at = now()
WHERE learned_preferences.status IN ('active', 'confirmed');

-- name: ListLearnedPreferences :many
SELECT id, user_id, preference_key, confidence, score, evidence_count,
    explanation_code, status, rules_version, first_observed_at, last_observed_at, updated_at
FROM learned_preferences WHERE user_id = $1 ORDER BY score DESC, preference_key;

-- name: GetLearnedPreference :one
SELECT id, user_id, preference_key, confidence, score, evidence_count,
    explanation_code, status, rules_version, first_observed_at, last_observed_at, updated_at
FROM learned_preferences WHERE id = $1;

-- name: UpdateLearnedPreferenceStatus :exec
UPDATE learned_preferences SET status = $2, updated_at = now() WHERE id = $1;

-- name: DeleteLearnedPreferences :exec
DELETE FROM learned_preferences WHERE user_id = $1 AND status IN ('active', 'confirmed');

-- name: SetHistoryLearning :exec
UPDATE traveler_preferences SET history_learning_enabled = $2, updated_at = now() WHERE user_id = $1;

-- name: GetHistoryLearning :one
SELECT history_learning_enabled FROM traveler_preferences WHERE user_id = $1;

-- name: UpsertTripFeedback :one
INSERT INTO trip_feedback (trip_id, user_id, positives, negatives, pace, travel_burden, return_intent, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (trip_id) DO UPDATE SET
    positives = EXCLUDED.positives,
    negatives = EXCLUDED.negatives,
    pace = EXCLUDED.pace,
    travel_burden = EXCLUDED.travel_burden,
    return_intent = EXCLUDED.return_intent,
    updated_at = now()
RETURNING trip_id, user_id, positives, negatives, pace, travel_burden, return_intent, submitted_at, updated_at;

-- name: GetTripFeedback :one
SELECT trip_id, user_id, positives, negatives, pace, travel_burden, return_intent, submitted_at, updated_at
FROM trip_feedback WHERE trip_id = $1;
