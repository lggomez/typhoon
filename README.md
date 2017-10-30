# typhoon
Typhoon is a golang static analyzer made in go. Its goal is to target similar string literals in order to be able to identify typos

Yes, the name is a poor attempt at a word play with typo

## Usage (command line args)
- **dir**: string - Path containing the source to analyze. If none, will use os.Getwd()
- **dist**: int - Levenshtein-Damerau distance threshold (default 2)
