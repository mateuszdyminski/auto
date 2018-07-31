package search

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/mateuszdyminski/auto/ingress/model"
	"github.com/olivere/elastic"
)

type Finder struct {
	queryString string
	from        time.Time
	to          time.Time
	skip        int
	size        int
	sort        []string
}

// NewFinder creates a new finder for logs.
// Use the funcs to set up filters and search properties,
// then call Find to execute.
func NewFinder() *Finder {
	return &Finder{}
}

// Query searches the results by the given string query.
func (f *Finder) Query(queryString string) *Finder {
	f.queryString = queryString
	return f
}

// From filters the results which occur after specified time.
func (f *Finder) From(from time.Time) *Finder {
	f.from = from
	return f
}

// To filters the results which occur before specified time.
func (f *Finder) To(to time.Time) *Finder {
	f.to = to
	return f
}

// Skip specifies the number of items to skip in pagination.
func (f *Finder) Skip(skip int) *Finder {
	f.skip = skip
	return f
}

// Size specifies the number of items to return in pagination.
func (f *Finder) Size(size int) *Finder {
	f.size = size
	return f
}

// Sort specifies one or more sort orders.
// Use a dash (-) to make the sort order descending.
// Example: "name" or "-year".
func (f *Finder) Sort(sort ...string) *Finder {
	if f.sort == nil {
		f.sort = make([]string, 0)
	}
	f.sort = append(f.sort, sort...)
	return f
}

// query sets up the query in the search service.
func (f *Finder) query(service *elastic.SearchService) *elastic.SearchService {
	if f.queryString == "" && f.from.IsZero() && f.to.IsZero() {
		service = service.Query(elastic.NewMatchAllQuery())
		return service
	}

	q := elastic.NewBoolQuery()
	if f.queryString != "" {
		q = q.Must(elastic.NewQueryStringQuery(f.queryString).Field("msg"))
	}
	if !f.from.IsZero() {
		q = q.Must(elastic.NewRangeQuery("time").Gte(f.from))
	}
	if !f.to.IsZero() {
		q = q.Must(elastic.NewRangeQuery("time").Lte(f.to))
	}

	service = service.Query(q)
	return service
}

// paginate sets up pagination in the service.
func (f *Finder) paginate(service *elastic.SearchService) *elastic.SearchService {
	if f.skip > 0 {
		service = service.From(f.skip)
	}
	if f.size > 0 {
		service = service.Size(f.size)
	}
	return service
}

// sorting applies sorting to the service.
func (f *Finder) sorting(service *elastic.SearchService) *elastic.SearchService {
	if len(f.sort) == 0 {
		// Sort by score by default
		service = service.Sort("_score", false)
		return service
	}

	// Sort by fields; prefix of "-" means: descending sort order.
	for _, s := range f.sort {
		s = strings.TrimSpace(s)

		var field string
		var asc bool

		if strings.HasPrefix(s, "-") {
			field = s[1:]
			asc = false
		} else {
			field = s
			asc = true
		}

		// Maybe check for permitted fields to sort

		service = service.Sort(field, asc)
	}
	return service
}

func (f *Finder) Find(client *elastic.Client) (*Response, error) {
	// Create service and use query, aggregations, sort, filter, pagination funcs
	search := client.Search().Index("flights").Type("flight")
	search = f.query(search)
	search = f.sorting(search)
	search = f.paginate(search)

	// Execute query
	searchResult, err := search.Do(context.Background())
	if err != nil {
		return nil, err
	}

	response := Response{}
	logs := make([]model.FlightCrash, 0)
	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits != nil {
		log.Printf("for (query: %s, from: %v, to: %v, size: %d, skip: %d) found a total of %d logs\n", f.queryString, f.from, f.to, f.size, f.skip, searchResult.Hits.TotalHits)

		response.Total = searchResult.Hits.TotalHits

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			var l model.FlightCrash
			err := json.Unmarshal(*hit.Source, &l)
			if err != nil {
				return nil, err
			}

			if hit.Score != nil {
				l.Score = hit.Score
			}

			l.ID = hit.Id
			logs = append(logs, l)
		}
	}
	response.Data = logs

	return &response, nil
}
