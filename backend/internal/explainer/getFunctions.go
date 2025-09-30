package explainer

// function are out of scope for now.
// instead i will chunk the file and merge the responses.

type Pattern struct {
	Kind string
	Regex string
}

type LangPatterns struct {
	Patterns []Pattern
}

var Registry = map[string]LangPatterns{
	"js": {
		Patterns: []Pattern{
			// function foo(...) { ... }
			{
				Kind: "function",
				Regex: `(?m)^\s*function\s*\(?`,
			},
		},
	},	
}

func GetFunctions(fileExtension string, fileContent string) {
	
}
