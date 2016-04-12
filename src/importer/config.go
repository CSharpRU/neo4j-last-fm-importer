package importer

type Config struct {
	Neo4j  struct {
		       Scheme   string
		       Host     string
		       Port     uint16
		       Username string
		       Password string
	       }
	LastFm struct {
		       Key     string
		       Secret  string
		       Workers int
		       Pages   int
	       }
}

var AppConfig Config