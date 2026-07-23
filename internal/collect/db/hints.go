package db

// Engine identifiers for discovered databases.
const (
	EnginePostgres      = "postgres"
	EngineRedis         = "redis"
	EngineMongoDB       = "mongodb"
	EngineMySQL         = "mysql"
	EngineQdrant        = "qdrant"
	EngineElasticsearch = "elasticsearch"
)

type engineHint struct {
	engine        string
	imagePatterns []string
	namePatterns  []string
	processNames  []string
	cmdPatterns   []string
	ports         []uint16
}

var engineHints = []engineHint{
	{
		engine:        EnginePostgres,
		imagePatterns: []string{"postgres", "postgresql", "timescale"},
		namePatterns:  []string{"postgres", "pg", "timescale"},
		processNames:  []string{"postgres"},
		ports:         []uint16{5432},
	},
	{
		engine:        EngineRedis,
		imagePatterns: []string{"redis"},
		namePatterns:  []string{"redis"},
		processNames:  []string{"redis-server", "redis"},
		ports:         []uint16{6379},
	},
	{
		engine:        EngineMongoDB,
		imagePatterns: []string{"mongo"},
		namePatterns:  []string{"mongo"},
		processNames:  []string{"mongod"},
		ports:         []uint16{27017},
	},
	{
		engine:        EngineMySQL,
		imagePatterns: []string{"mysql", "mariadb", "percona"},
		namePatterns:  []string{"mysql", "mariadb"},
		processNames:  []string{"mysqld", "mariadbd"},
		ports:         []uint16{3306},
	},
	{
		engine:        EngineQdrant,
		imagePatterns: []string{"qdrant"},
		namePatterns:  []string{"qdrant"},
		processNames:  []string{"qdrant"},
		ports:         []uint16{6333, 6334},
	},
	{
		engine:        EngineElasticsearch,
		imagePatterns: []string{"elasticsearch", "opensearch"},
		namePatterns:  []string{"elasticsearch", "opensearch"},
		processNames:  []string{"java"},
		cmdPatterns:   []string{"elasticsearch", "opensearch"},
		ports:         []uint16{9200, 9300},
	},
}

var defaultPortEngines = map[uint16]string{
	5432:  EnginePostgres,
	6379:  EngineRedis,
	27017: EngineMongoDB,
	3306:  EngineMySQL,
	6333:  EngineQdrant,
	9200:  EngineElasticsearch,
}

func matchEngineHint(image, name string) *engineHint {
	for i := range engineHints {
		h := &engineHints[i]
		if matchesAny(image, h.imagePatterns) || matchesAny(name, h.namePatterns) {
			return h
		}
	}
	return nil
}

func matchProcessEngine(name, cmdline string) *engineHint {
	name = normalizeProcName(name)
	cmdline = lower(cmdline)
	for i := range engineHints {
		h := &engineHints[i]
		for _, pn := range h.processNames {
			if name == pn {
				if len(h.cmdPatterns) == 0 || matchesAny(cmdline, h.cmdPatterns) {
					return h
				}
			}
		}
		if matchesAny(cmdline, h.cmdPatterns) {
			return h
		}
	}
	return nil
}

func displayName(engine string) string {
	switch engine {
	case EnginePostgres:
		return "PostgreSQL"
	case EngineRedis:
		return "Redis"
	case EngineMongoDB:
		return "MongoDB"
	case EngineMySQL:
		return "MySQL"
	case EngineQdrant:
		return "Qdrant"
	case EngineElasticsearch:
		return "Elasticsearch"
	default:
		return engine
	}
}

func matchesAny(s string, patterns []string) bool {
	s = lower(s)
	for _, p := range patterns {
		if p != "" && containsFold(s, p) {
			return true
		}
	}
	return false
}

func normalizeProcName(name string) string {
	name = lower(name)
	if len(name) > 4 && name[len(name)-4:] == ".exe" {
		name = name[:len(name)-4]
	}
	return name
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsFold(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && indexFold(s, sub) >= 0
}

func indexFold(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if equalFoldAt(s, i, sub) {
			return i
		}
	}
	return -1
}

func equalFoldAt(s string, i int, sub string) bool {
	for j := 0; j < len(sub); j++ {
		a, b := s[i+j], sub[j]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}
