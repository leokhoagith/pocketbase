package daos

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/list"
)

// MaxExpandDepth specifies the max allowed nested expand depth path.
const MaxExpandDepth = 6

// ExpandFetchFunc defines the function that is used to fetch the expanded relation records.
type ExpandFetchFunc func(relCollection *models.Collection, relIds []string) ([]*models.Record, error)

// ExpandRecord expands the relations of a single Record model.
func (dao *Dao) ExpandRecord(record *models.Record, expands []string, fetchFunc ExpandFetchFunc) error {
	return dao.ExpandRecords([]*models.Record{record}, expands, fetchFunc)
}

// ExpandRecords expands the relations of the provided Record models list.
func (dao *Dao) ExpandRecords(records []*models.Record, expands []string, fetchFunc ExpandFetchFunc) error {
	normalized := normalizeExpands(expands)

	for _, expand := range normalized {
		if err := dao.expandRecords(records, expand, fetchFunc, 1); err != nil {
			return err
		}
	}

	return nil
}

// notes:
// - fetchFunc must be non-nil func
// - all records are expected to be from the same collection
// - if MaxExpandDepth is reached, the function returns nil ignoring the remaining expand path
func (dao *Dao) expandRecords(records []*models.Record, expandPath string, fetchFunc ExpandFetchFunc, recursionLevel int) error {
	if fetchFunc == nil {
		return errors.New("Relation records fetchFunc is not set.")
	}

	if expandPath == "" || recursionLevel > MaxExpandDepth || len(records) == 0 {
		return nil
	}

	parts := strings.SplitN(expandPath, ".", 2)

	// extract the relation field (if exist)
	mainCollection := records[0].Collection()
	relField := mainCollection.Schema.GetFieldByName(parts[0])
	if relField == nil {
		return fmt.Errorf("Couldn't find field %q in collection %q.", parts[0], mainCollection.Name)
	}
	relField.InitOptions()
	relFieldOptions, _ := relField.Options.(*schema.RelationOptions)

	relCollection, err := dao.FindCollectionByNameOrId(relFieldOptions.CollectionId)
	if err != nil {
		return fmt.Errorf("Couldn't find collection %q.", relFieldOptions.CollectionId)
	}

	// extract the id of the relations to expand
	relIds := []string{}
	for _, record := range records {
		relIds = append(relIds, record.GetStringSliceDataValue(relField.Name)...)
	}

	// fetch rels
	rels, relsErr := fetchFunc(relCollection, relIds)
	if relsErr != nil {
		return relsErr
	}

	// expand nested fields
	if len(parts) > 1 {
		err := dao.expandRecords(rels, parts[1], fetchFunc, recursionLevel+1)
		if err != nil {
			return err
		}
	}

	// reindex with the rel id
	indexedRels := map[string]*models.Record{}
	for _, rel := range rels {
		indexedRels[rel.GetId()] = rel
	}

	for _, model := range records {
		relIds := model.GetStringSliceDataValue(relField.Name)

		validRels := []*models.Record{}
		for _, id := range relIds {
			if rel, ok := indexedRels[id]; ok {
				validRels = append(validRels, rel)
			}
		}

		if len(validRels) == 0 {
			continue // no valid relations
		}

		expandData := model.GetExpand()

		// normalize and set the expanded relations
		if relFieldOptions.MaxSelect == 1 {
			expandData[relField.Name] = validRels[0]
		} else {
			expandData[relField.Name] = validRels
		}
		model.SetExpand(expandData)
	}

	return nil
}

// normalizeExpands normalizes expand strings and merges self containing paths
// (eg. ["a.b.c", "a.b", "   test  ", "  ", "test"] -> ["a.b.c", "test"]).
func normalizeExpands(paths []string) []string {
	result := []string{}

	// normalize paths
	normalized := []string{}
	for _, p := range paths {
		p := strings.ReplaceAll(p, " ", "") // replace spaces
		p = strings.Trim(p, ".")            // trim incomplete paths
		if p == "" {
			continue
		}
		normalized = append(normalized, p)
	}

	// merge containing paths
	for i, p1 := range normalized {
		var skip bool
		for j, p2 := range normalized {
			if i == j {
				continue
			}
			if strings.HasPrefix(p2, p1+".") {
				// skip because there is more detailed expand path
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, p1)
		}
	}

	return list.ToUniqueStringSlice(result)
}
