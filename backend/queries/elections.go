package queries

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/oriolf/bella-ciao/backend/models"
	"github.com/pkg/errors"
)

func GetElections(db *sql.DB, onlyPublic bool) ([]models.Election, error) {
	results, err := queryDB(db, scanElection, `
		SELECT id, name, date_start, date_end, count_type, max_candidates, min_candidates, public 
		FROM elections WHERE public OR public = $1 ORDER BY date_start ASC;`, onlyPublic)
	if err != nil {
		return nil, errors.Wrap(err, "error querying elections")
	}

	var electionIDs []int
	electionsMap := make(map[int]models.Election)
	for _, x := range results {
		e, _ := x.(models.Election)
		if e.ID == 0 {
			continue
		}
		electionsMap[e.ID] = e
		electionIDs = append(electionIDs, e.ID)
	}

	results, err = queryDB(db, scanCandidate, `
		SELECT id, election_id, name, presentation, image FROM candidates WHERE election_id=ANY($1) ORDER BY random();`,
		pq.Array(electionIDs))
	if err != nil {
		return nil, errors.Wrap(err, "error querying candidates")
	}

	for _, x := range results {
		c, _ := x.(models.Candidate)
		e := electionsMap[c.ElectionID]
		if c.ID == 0 || e.ID == 0 {
			continue
		}
		e.Candidates = append(e.Candidates, c)
		electionsMap[c.ElectionID] = e
	}

	var elections []models.Election
	for _, id := range electionIDs {
		elections = append(elections, electionsMap[id])
	}

	return elections, nil
}

func scanElection(rows *sql.Rows) (interface{}, error) {
	var e models.Election
	err := rows.Scan(&e.ID, &e.Name, &e.Start, &e.End, &e.CountType, &e.MaxCandidates, &e.MinCandidates, &e.Public)
	return e, err
}

func scanCandidate(rows *sql.Rows) (interface{}, error) {
	var c models.Candidate
	err := rows.Scan(&c.ID, &c.ElectionID, &c.Name, &c.Presentation, &c.Image)
	return c, err
}
