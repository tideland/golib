// Tideland Go Library - Audit
//
// Copyright (C) 2013-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package audit

//--------------------
// IMPORTS
//--------------------

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

//--------------------
// HELPERS
//--------------------

// SimpleRand returns a random number generator with a source using
// the the current time as seed. It's not the best random, but ok to
// generate test data.
func SimpleRand() *rand.Rand {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	return rand.New(source)
}

// FixedRand returns a random number generator with a fixed source
// so that tests using the generate functions can be repeated with
// the same result.
func FixedRand() *rand.Rand {
	source := rand.NewSource(42)
	return rand.New(source)
}

// ToUpperFirst returns the passed string with the first rune
// converted to uppercase.
func ToUpperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// BuildEMail creates an e-mail address out of first and last
// name and the domain.
func BuildEMail(first, last, domain string) string {
	valid := make(map[rune]bool)
	for _, r := range "abcdefghijklmnopqrstuvwxyz0123456789-" {
		valid[r] = true
	}
	name := func(in string) string {
		out := []rune{}
		for _, r := range strings.ToLower(in) {
			if valid[r] {
				out = append(out, r)
			}
		}
		return string(out)
	}
	return name(first) + "." + name(last) + "@" + domain
}

// BuildTime returns the current time plus or minus the passed
// offset formatted as string and as Time. The returned time is
// the parsed formatted one to avoid parsing troubles in tests.
func BuildTime(layout string, offset time.Duration) (string, time.Time) {
	t := time.Now().Add(offset)
	ts := t.Format(layout)
	tp, err := time.Parse(layout, ts)
	if err != nil {
		panic("cannot build time: " + err.Error())
	}
	return ts, tp
}

//--------------------
// GENERATOR
//--------------------

// Generator is responsible for generating different random data
// based on a random number generator.
type Generator struct {
	rand *rand.Rand
}

// NewGenerator returns a new generator using the passed random number
// generator.
func NewGenerator(rand *rand.Rand) *Generator {
	return &Generator{rand}
}

// Int generates an int between lo and hi including
// those values.
func (g *Generator) Int(lo, hi int) int {
	if lo == hi {
		return lo
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	n := g.rand.Intn(hi - lo + 1)
	return lo + n
}

// Ints generates a slice of random ints.
func (g *Generator) Ints(lo, hi, count int) []int {
	ints := make([]int, count)
	for i := 0; i < count; i++ {
		ints[i] = g.Int(lo, hi)
	}
	return ints
}

// Percent generates an int between 0 and 100.
func (g *Generator) Percent() int {
	return g.Int(0, 100)
}

// FlipCoin returns true if the internal generated percentage is
// equal or greater than the passed percentage.
func (g *Generator) FlipCoin(percent int) bool {
	switch {
	case percent > 100:
		percent = 100
	case percent < 0:
		percent = 0
	}
	return g.Percent() >= percent
}

// OneByteOf returns one of the passed bytes.
func (g *Generator) OneByteOf(values ...byte) byte {
	i := g.Int(0, len(values)-1)
	return values[i]
}

// OneRuneOf returns one of the runes of the passed string.
func (g *Generator) OneRuneOf(values string) rune {
	runes := []rune(values)
	i := g.Int(0, len(runes)-1)
	return runes[i]
}

// OneIntOf returns one of the passed ints.
func (g *Generator) OneIntOf(values ...int) int {
	i := g.Int(0, len(values)-1)
	return values[i]
}

// OneStringOf returns one of the passed strings.
func (g *Generator) OneStringOf(values ...string) string {
	i := g.Int(0, len(values)-1)
	return values[i]
}

// OneDurationOf returns one of the passed durations.
func (g *Generator) OneDurationOf(values ...time.Duration) time.Duration {
	i := g.Int(0, len(values)-1)
	return values[i]
}

// Word generates a random word.
func (g *Generator) Word() string {
	return g.OneStringOf(words...)
}

// Words generates a slice of random words
func (g *Generator) Words(count int) []string {
	words := make([]string, count)
	for i := 0; i < count; i++ {
		words[i] = g.Word()
	}
	return words
}

// LimitedWord generates a random word with a length between
// lo and hi.
func (g *Generator) LimitedWord(lo, hi int) string {
	length := g.Int(lo, hi)
	if length < MinWordLen {
		length = MinWordLen
	}
	if length > MaxWordLen {
		length = MaxWordLen
	}
	// Start anywhere in the list.
	pos := g.Int(0, wordsLen)
	for {
		if pos >= wordsLen {
			pos = 0
		}
		if len(words[pos]) == length {
			return words[pos]
		}
		pos++
	}
}

// Pattern generates a string based on a pattern. Here different
// escape chars are replaced by according random chars while all
// others are left as they are. Escape chars start with a caret (^)
// followed by specializer. Those are:
//
//   - ^ for a caret
//   - 0 for a number between 0 and 9
//   - 1 for a number between 1 and 9
//   - o for an octal number
//   - h for a hexadecimal number (lower-case)
//   - H for a hexadecimal number (upper-case)
//   - a for any char between a and z
//   - A for any char between A and Z
//   - c for a consonant (lower-case)
//   - C for a consonant (upper-case)
//   - v for a vowel (lower-case)
//   - V for a vowel (upper-case)
func (g *Generator) Pattern(pattern string) string {
	result := []rune{}
	escaped := false
	for _, pr := range pattern {
		if !escaped {
			if pr == '^' {
				escaped = true
			} else {
				result = append(result, pr)
			}
			continue
		}
		// Escaped mode.
		ar := pr
		switch pr {
		case '0':
			ar = g.OneRuneOf("0123456789")
		case '1':
			ar = g.OneRuneOf("123456789")
		case 'o', 'O':
			ar = g.OneRuneOf("01234567")
		case 'h':
			ar = g.OneRuneOf("0123456789abcdef")
		case 'H':
			ar = g.OneRuneOf("0123456789ABCDEF")
		case 'a':
			ar = g.OneRuneOf("abcdefghijklmnopqrstuvwxyz")
		case 'A':
			ar = g.OneRuneOf("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		case 'c':
			ar = g.OneRuneOf("bcdfghjklmnpqrstvwxyz")
		case 'C':
			ar = g.OneRuneOf("BCDFGHJKLMNPQRSTVWXYZ")
		case 'v':
			ar = g.OneRuneOf("aeiou")
		case 'V':
			ar = g.OneRuneOf("AEIOU")
		case 'z':
			ar = g.OneRuneOf("abcdefghijklmnopqrstuvwxyz0123456789")
		case 'Z':
			ar = g.OneRuneOf("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		}
		result = append(result, ar)
		escaped = false
	}
	return string(result)
}

// Sentence generates a sentence between 2 and 15 words
// and possibly containing commas.
func (g *Generator) Sentence() string {
	count := g.Int(2, 15)
	words := g.Words(count)
	words[0] = ToUpperFirst(words[0])
	for i := 2; i < count-1; i++ {
		if g.FlipCoin(80) {
			words[i] += ","
		}
	}
	return strings.Join(words, " ") + "."
}

// Paragraph generates a paragraph between 2 and 10 sentences.
func (g *Generator) Paragraph() string {
	count := g.Int(2, 10)
	sentences := make([]string, count)
	for i := 0; i < count; i++ {
		sentences[i] = g.Sentence()
	}
	return strings.Join(sentences, " ")
}

// Name generates a male or female name consisting out of first,
// middle and last name.
func (g *Generator) Name() (first, middle, last string) {
	if g.FlipCoin(50) {
		return g.FemaleName()
	}
	return g.MaleName()
}

// MaleName generates a male name consisting out of first, middle
// and last name.
func (g *Generator) MaleName() (first, middle, last string) {
	first = g.OneStringOf(maleFirstNames...)
	middle = g.OneStringOf(maleFirstNames...)
	if g.FlipCoin(80) {
		first += "-" + g.OneStringOf(maleFirstNames...)
	} else if g.FlipCoin(80) {
		middle += "-" + g.OneStringOf(maleFirstNames...)
	}
	last = g.OneStringOf(lastNames...)
	return
}

// FemaleName generates a female name consisting out of first, middle
// and last name.
func (g *Generator) FemaleName() (first, middle, last string) {
	first = g.OneStringOf(femaleFirstNames...)
	middle = g.OneStringOf(femaleFirstNames...)
	if g.FlipCoin(80) {
		first += "-" + g.OneStringOf(femaleFirstNames...)
	} else if g.FlipCoin(80) {
		middle += "-" + g.OneStringOf(femaleFirstNames...)
	}
	last = g.OneStringOf(lastNames...)
	return
}

// Domain generates domain out of name and top level domain.
func (g *Generator) Domain() string {
	tld := g.OneStringOf(topLevelDomains...)
	if g.FlipCoin(80) {
		return g.LimitedWord(3, 5) + "-" + g.LimitedWord(3, 5) + "." + tld
	}
	return g.LimitedWord(3, 10) + "." + tld
}

// URL generates a http, https or ftp URL, some of the leading
// to a file.
func (g *Generator) URL() string {
	part := func() string {
		return g.LimitedWord(2, 8)
	}
	start := g.OneStringOf("http://www.", "http://blog.", "https://www.", "ftp://")
	ext := g.OneStringOf("html", "php", "jpg", "mp3", "txt")
	variant := g.Percent()
	switch {
	case variant < 20:
		return start + part() + "." + g.Domain() + "/" + part() + "." + ext
	case variant > 80:
		return start + part() + "." + g.Domain() + "/" + part() + "/" + part() + "." + ext
	default:
		return start + part() + "." + g.Domain()
	}
}

// EMail returns a random e-mail address.
func (g *Generator) EMail() string {
	if g.FlipCoin(50) {
		first, _, last := g.MaleName()
		return BuildEMail(first, last, g.Domain())
	}
	first, _, last := g.FemaleName()
	return BuildEMail(first, last, g.Domain())
}

// Duration generates a duration between lo and hi including
// those values.
func (g *Generator) Duration(lo, hi time.Duration) time.Duration {
	if lo == hi {
		return lo
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	n := g.rand.Int63n(int64(hi) - int64(lo) + 1)
	return lo + time.Duration(n)
}

// SleepOneOf chooses randomely one of the passed durations
// and lets the goroutine sleep for this time.
func (g *Generator) SleepOneOf(sleeps ...time.Duration) time.Duration {
	sleep := g.OneDurationOf(sleeps...)
	time.Sleep(sleep)
	return sleep
}

// Time generates a time between the given one and that time
// plus the given duration. The result will have the passed
// location.
func (g *Generator) Time(loc *time.Location, base time.Time, dur time.Duration) time.Time {
	base = base.UTC()
	return base.Add(g.Duration(0, dur)).In(loc)
}

//--------------------
// GENERATOR DATA
//--------------------

// words is a list of words based on lorem ipsum and own extensions.
var words = []string{
	"a", "ac", "accumsan", "accusam", "accusantium", "ad", "adipiscing",
	"alias", "aliquam", "aliquet", "aliquip", "aliquyam", "amet", "aenean",
	"ante", "aperiam", "arcu", "assum", "at", "auctor", "augue", "aut", "autem",
	"bibendum", "blandit", "blanditiis",
	"clita", "commodo", "condimentum", "congue", "consectetuer", "consequat",
	"consequatur", "consequuntur", "consetetur", "convallis", "cras", "cubilia",
	"culpa", "cum", "curabitur", "curae", "cursus",
	"dapibus", "delectus", "delenit", "diam", "dictum", "dictumst", "dignissim", "dis",
	"dolor", "dolore", "dolores", "doloremque", "doming", "donec", "dui", "duis", "duo",
	"ea", "eaque", "earum", "egestas", "eget", "eirmod", "eleifend", "elementum", "elit",
	"elitr", "enim", "eos", "erat", "eros", "errare", "error", "esse", "est", "et", "etiam",
	"eu", "euismod", "eum", "ex", "exerci", "exercitationem",
	"facer", "facilisi", "facilisis", "fames", "faucibus", "felis", "fermentum",
	"feugait", "feugiat", "fringilla", "fuga", "fusce",
	"gravida", "gubergren",
	"habitant", "habitasse", "hac", "harum", "hendrerit", "hic",
	"iaculis", "id", "illum", "illo", "imperdiet", "in", "integer", "interdum",
	"invidunt", "ipsa", "ipsum", "iriure", "iusto",
	"justo",
	"kasd", "kuga",
	"labore", "lacinia", "lacus", "laoreet", "laudantium", "lectus", "leo", "liber",
	"libero", "ligula", "lobortis", "laboriosam", "lorem", "luctus", "luptatum",
	"maecenas", "magna", "magni", "magnis", "malesuada", "massa", "mattis", "mauris",
	"mazim", "mea", "metus", "mi", "minim", "molestie", "mollis", "montes", "morbi", "mus",
	"nam", "nascetur", "natoque", "nec", "neque", "nesciunt", "netus", "nibh", "nihil",
	"nisi", "nisl", "no", "nobis", "non", "nonummy", "nonumy", "nostrud", "nulla",
	"nullam", "nunc",
	"odio", "odit", "officia", "option", "orci", "ornare",
	"parturient", "pede", "pellentesque", "penatibus", "perfendis", "perspiciatis",
	"pharetra", "phasellus", "placerat", "platea", "porta", "porttitor", "possim",
	"posuere", "praesent", "praesentium", "pretium", "primis", "proin", "pulvinar",
	"purus",
	"quam", "qui", "quia", "quis", "quisque", "quod",
	"rebum", "rhoncus", "ridiculus", "risus", "rutrum",
	"sadipscing", "sagittis", "sanctus", "sapien", "scelerisque", "sea", "sed",
	"sem", "semper", "senectus", "sit", "sociis", "sodales", "sollicitudin", "soluta",
	"stet", "suscipit", "suspendisse",
	"takimata", "tation", "te", "tellus", "tempor", "tempora", "temporibus", "tempus",
	"tincidunt", "tortor", "totam", "tristique", "turpis",
	"ullam", "ullamcorper", "ultrices", "ultricies", "urna", "ut",
	"varius", "vehicula", "vel", "velit", "venenatis", "veniam", "vero", "vestibulum",
	"vitae", "vivamus", "viverra", "voluptua", "volutpat", "voluptatem", "vulputate",
	"voluptatem",
	"wisi", "wiskaleborium",
	"xantippe", "xeon",
	"yodet", "yggdrasil",
	"zypres", "zyril",
}

// wordsLen is the length of the word list.
var wordsLen = len(words)

const (
	// MinWordLen is the length of the shortest word.
	MinWordLen = 1

	// MaxWordLen is the length of the longest word.
	MaxWordLen = 14
)

// maleFirstNames is a list of popular male first names.
var maleFirstNames = []string{
	"Jacob", "Michael", "Joshua", "Matthew", "Ethan", "Andrew", "Daniel",
	"Anthony", "Christopher", "Joseph", "William", "Alexander", "Ryan", "David",
	"Nicholas", "Tyler", "James", "John", "Jonathan", "Nathan", "Samuel",
	"Christian", "Noah", "Dylan", "Benjamin", "Logan", "Brandon", "Gabriel",
	"Zachary", "Jose", "Elijah", "Angel", "Kevin", "Jack", "Caleb", "Justin",
	"Austin", "Evan", "Robert", "Thomas", "Luke", "Mason", "Aidan", "Jackson",
	"Isaiah", "Jordan", "Gavin", "Connor", "Aiden", "Isaac", "Jason", "Cameron",
	"Hunter", "Jayden", "Juan", "Charles", "Aaron", "Lucas", "Luis", "Owen",
	"Landon", "Diego", "Brian", "Adam", "Adrian", "Kyle", "Eric", "Ian", "Nathaniel",
	"Carlos", "Alex", "Bryan", "Jesus", "Julian", "Sean", "Carter", "Hayden",
	"Jeremiah", "Cole", "Brayden", "Wyatt", "Chase", "Steven", "Timothy", "Dominic",
	"Sebastian", "Xavier", "Jaden", "Jesse", "Devin", "Seth", "Antonio", "Richard",
	"Miguel", "Colin", "Cody", "Alejandro", "Caden", "Blake", "Carson",
}

// maleFirstNames is a list of popular female first names.
var femaleFirstNames = []string{
	"Emily", "Emma", "Madison", "Abigail", "Olivia", "Isabella", "Hannah",
	"Samantha", "Ava", "Ashley", "Sophia", "Elizabeth", "Alexis", "Grace",
	"Sarah", "Alyssa", "Mia", "Natalie", "Chloe", "Brianna", "Lauren", "Ella",
	"Anna", "Taylor", "Kayla", "Hailey", "Jessica", "Victoria", "Jasmine", "Sydney",
	"Julia", "Destiny", "Morgan", "Kaitlyn", "Savannah", "Katherine", "Alexandra",
	"Rachel", "Lily", "Megan", "Kaylee", "Jennifer", "Angelina", "Makayla", "Allison",
	"Brooke", "Maria", "Trinity", "Lillian", "Mackenzie", "Faith", "Sofia", "Riley",
	"Haley", "Gabrielle", "Nicole", "Kylie", "Katelyn", "Zoe", "Paige", "Gabriella",
	"Jenna", "Kimberly", "Stephanie", "Alexa", "Avery", "Andrea", "Leah", "Madeline",
	"Nevaeh", "Evelyn", "Maya", "Mary", "Michelle", "Jada", "Sara", "Audrey",
	"Brooklyn", "Vanessa", "Amanda", "Ariana", "Rebecca", "Caroline", "Amelia",
	"Mariah", "Jordan", "Jocelyn", "Arianna", "Isabel", "Marissa", "Autumn", "Melanie",
	"Aaliyah", "Gracie", "Claire", "Isabelle", "Molly", "Mya", "Diana", "Katie",
}

// lastNames is a list of popular last names.
var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia",
	"Rodriguez", "Wilson", "Martinez", "Anderson", "Taylor", "Thomas", "Hernandez",
	"Moore", "Martin", "Jackson", "Thompson", "White", "Lopez", "Lee", "Gonzalez",
	"Harris", "Clark", "Lewis", "Robinson", "Walker", "Perez", "Hall", "Young",
	"Allen", "Sanchez", "Wright", "King", "Scott", "Green", "Baker", "Adams", "Nelson",
	"Hill", "Ramirez", "Campbell", "Mitchell", "Roberts", "Carter", "Phillips", "Evans",
	"Turner", "Torres", "Parker", "Collins", "Edwards", "Stewart", "Flores", "Morris",
	"Nguyen", "Murphy", "Rivera", "Cook", "Rogers", "Morgan", "Peterson", "Cooper",
	"Reed", "Bailey", "Bell", "Gomez", "Kelly", "Howard", "Ward", "Cox", "Diaz",
	"Richardson", "Wood", "Watson", "Brooks", "Bennett", "Gray", "James", "Reyes",
	"Cruz", "Hughes", "Price", "Myers", "Long", "Foster", "Sanders", "Ross", "Morales",
	"Powell", "Sullivan", "Russell", "Ortiz", "Jenkins", "Gutierrez", "Perry", "Butler",
	"Barnes", "Fisher", "Henderson", "Coleman", "Simmons", "Patterson", "Jordan",
	"Reynolds", "Hamilton", "Graham", "Kim", "Gonzales", "Alexander", "Ramos", "Wallace",
	"Griffin", "West", "Cole", "Hayes", "Chavez", "Gibson", "Bryant", "Ellis", "Stevens",
	"Murray", "Ford", "Marshall", "Owens", "McDonald", "Harrison", "Ruiz", "Kennedy",
	"Wells", "Alvarez", "Woods", "Mendoza", "Castillo", "Olson", "Webb", "Washington",
	"Tucker", "Freeman", "Burns", "Henry", "Vasquez", "Snyder", "Simpson", "Crawford",
	"Jimenez", "Porter", "Mason", "Shaw", "Gordon", "Wagner", "Hunter", "Romero",
	"Hicks", "Dixon", "Hunt", "Palmer", "Robertson", "Black", "Holmes", "Stone",
	"Meyer", "Boyd", "Mills", "Warren", "Fox", "Rose", "Rice", "Moreno", "Schmidt",
	"Patel", "Ferguson", "Nichols", "Herrera", "Medina", "Ryan", "Fernandez", "Weaver",
	"Daniels", "Stephens", "Gardner", "Payne", "Kelley", "Dunn", "Pierce", "Arnold",
	"Tran", "Spencer", "Peters", "Hawkins", "Grant", "Hansen", "Castro", "Hoffman",
	"Hart", "Elliott", "Cunningham", "Knight", "Bradley", "Carroll", "Hudson", "Duncan",
	"Armstrong", "Berry", "Andrews", "Johnston", "Ray", "Lane", "Riley", "Carpenter",
	"Perkins", "Aguilar", "Silva", "Richards", "Willis", "Matthews", "Chapman",
	"Lawrence", "Garza", "Vargas", "Watkins", "Wheeler", "Larson", "Carlson", "Harper",
	"George", "Greene", "Burke", "Guzman", "Morrison", "Munoz", "Jacobs", "Obrien",
	"Lawson", "Franklin", "Lynch", "Bishop", "Carr", "Salazar", "Austin", "Mendez",
	"Gilbert", "Jensen", "Williamson", "Montgomery", "Harvey", "Oliver", "Howell",
	"Dean", "Hanson", "Weber", "Garrett", "Sims", "Burton", "Fuller", "Soto", "McCoy",
	"Welch", "Chen", "Schultz", "Walters", "Reid", "Fields", "Walsh", "Little", "Fowler",
	"Bowman", "Davidson", "May", "Day", "Schneider", "Newman", "Brewer", "Lucas", "Holland",
	"Wong", "Banks", "Santos", "Curtis", "Pearson", "Delgado", "Valdez", "Pena", "Rios",
	"Douglas", "Sandoval", "Barrett", "Hopkins", "Keller", "Guerrero", "Stanley", "Bates",
	"Alvarado", "Beck", "Ortega", "Wade", "Estrada", "Contreras", "Barnett", "Caldwell",
	"Santiago", "Lambert", "Powers", "Chambers", "Nunez", "Craig", "Leonard", "Lowe", "Rhodes",
	"Byrd", "Gregory", "Shelton", "Frazier", "Becker", "Maldonado", "Fleming", "Vega",
	"Sutton", "Cohen", "Jennings", "Parks", "McDaniel", "Watts", "Barker", "Norris",
	"Vaughn", "Vazquez", "Holt", "Schwartz", "Steele", "Benson", "Neal", "Dominguez",
	"Horton", "Terry", "Wolfe", "Hale", "Lyons", "Graves", "Haynes", "Miles", "Park",
	"Warner", "Padilla", "Bush", "Thornton", "McCarthy", "Mann", "Zimmerman", "Erickson",
	"Fletcher", "McKinney", "Page", "Dawson", "Joseph", "Marquez", "Reeves", "Klein",
	"Espinoza", "Baldwin", "Moran", "Love", "Robbins", "Higgins", "Ball", "Cortez", "Le",
	"Griffith", "Bowen", "Sharp", "Cummings", "Ramsey", "Hardy", "Swanson", "Barber",
	"Acosta", "Luna", "Chandler", "Blair", "Daniel", "Cross", "Simon", "Dennis", "Oconnor",
	"Quinn", "Gross", "Navarro", "Moss", "Fitzgerald", "Doyle", "McLaughlin", "Rojas",
	"Rodgers", "Stevenson", "Singh", "Yang", "Figueroa", "Harmon", "Newton", "Paul",
	"Manning", "Garner", "McGee", "Reese", "Francis", "Burgess", "Adkins", "Goodman",
	"Curry", "Brady", "Christensen", "Potter", "Walton", "Goodwin", "Mullins", "Molina",
	"Webster", "Fischer", "Campos", "Avila", "Sherman", "Todd", "Chang", "Blake", "Malone",
	"Wolf", "Hodges", "Juarez", "Gill", "Farmer", "Hines", "Gallagher", "Duran", "Hubbard",
	"Cannon", "Miranda", "Wang", "Saunders", "Tate", "Mack", "Hammond", "Carrillo",
	"Townsend", "Wise", "Ingram", "Barton", "Mejia", "Ayala", "Schroeder", "Hampton",
	"Rowe", "Parsons", "Frank", "Waters", "Strickland", "Osborne", "Maxwell", "Chan",
	"Deleon", "Norman", "Harrington", "Casey", "Patton", "Logan", "Bowers", "Mueller",
	"Glover", "Floyd", "Hartman", "Buchanan", "Cobb", "French", "Kramer", "McCormick",
	"Clarke", "Tyler", "Gibbs", "Moody", "Conner", "Sparks", "McGuire", "Leon", "Bauer",
	"Norton", "Pope", "Flynn", "Hogan", "Robles", "Salinas", "Yates", "Lindsey", "Lloyd",
	"Marsh", "McBride", "Owen", "Solis", "Pham", "Lang", "Pratt", "Lara", "Brock", "Ballard",
	"Trujillo", "Shaffer", "Drake", "Roman", "Aguirre", "Morton", "Stokes", "Lamb", "Pacheco",
	"Patrick", "Cochran", "Shepherd", "Cain", "Burnett", "Hess", "Li", "Cervantes", "Olsen",
	"Briggs", "Ochoa", "Cabrera", "Velasquez", "Montoya", "Roth", "Meyers", "Cardenas", "Fuentes",
	"Weiss", "Hoover", "Wilkins", "Nicholson", "Underwood", "Short", "Carson", "Morrow", "Colon",
	"Holloway", "Summers", "Bryan", "Petersen", "McKenzie", "Serrano", "Wilcox", "Carey", "Clayton",
	"Poole", "Calderon", "Gallegos", "Greer", "Rivas", "Guerra", "Decker", "Collier", "Wall",
	"Whitaker", "Bass", "Flowers", "Davenport", "Conley", "Houston", "Huff", "Copeland", "Hood",
	"Monroe", "Massey", "Roberson", "Combs", "Franco", "Larsen", "Pittman", "Randall", "Skinner",
	"Wilkinson", "Kirby", "Cameron", "Bridges", "Anthony", "Richard", "Kirk", "Bruce", "Singleton",
	"Mathis", "Bradford", "Boone", "Abbott", "Charles", "Allison", "Sweeney", "Atkinson", "Horn",
	"Jefferson", "Rosales", "York", "Christian", "Phelps", "Farrell", "Castaneda", "Nash",
	"Dickerson", "Bond", "Wyatt", "Foley", "Chase", "Gates", "Vincent", "Mathews", "Hodge",
	"Garrison", "Trevino", "Villarreal", "Heath", "Dalton", "Valencia", "Callahan", "Hensley",
	"Atkins", "Huffman", "Roy", "Boyer", "Shields", "Lin", "Hancock", "Grimes", "Glenn", "Cline",
	"Delacruz", "Camacho", "Dillon", "Parrish", "O'Neill", "Melton", "Booth", "Kane", "Berg",
	"Harrell", "Pitts", "Savage", "Wiggins", "Brennan", "Salas", "Marks", "Russo", "Sawyer",
	"Baxter", "Golden", "Hutchinson", "Liu", "Walter", "McDowell", "Wiley", "Rich", "Humphrey",
	"Johns", "Koch", "Suarez", "Hobbs", "Beard", "Gilmore", "Ibarra", "Keith", "Macias", "Khan",
	"Andrade", "Ware", "Stephenson", "Henson", "Wilkerson", "Dyer", "McClure", "Blackwell",
	"Mercado", "Tanner", "Eaton", "Clay", "Barron", "Beasley", "O'Neal", "Preston", "Small",
	"Wu", "Zamora", "Macdonald", "Vance", "Snow", "McClain", "Stafford", "Orozco", "Barry",
	"English", "Shannon", "Kline", "Jacobson", "Woodard", "Huang", "Kemp", "Mosley", "Prince",
	"Merritt", "Hurst", "Villanueva", "Roach", "Nolan", "Lam", "Yoder", "McCullough", "Lester",
	"Santana", "Valenzuela", "Winters", "Barrera", "Leach", "Orr", "Berger", "McKee", "Strong",
	"Conway", "Stein", "Whitehead", "Bullock", "Escobar", "Knox", "Meadows", "Solomon", "Velez",
	"Odonnell", "Kerr", "Stout", "Blankenship", "Browning", "Kent", "Lozano", "Bartlett", "Pruitt",
	"Buck", "Barr", "Gaines", "Durham", "Gentry", "McIntyre", "Sloan", "Melendez", "Rocha", "Herman",
	"Sexton", "Moon", "Hendricks", "Rangel",
}

// topLevelDomains is a number of existing top level domains.
var topLevelDomains = []string{"asia", "at", "au", "biz", "ch", "cn", "com", "de", "es",
	"eu", "fr", "gr", "guru", "info", "it", "mobi", "name", "net", "org", "pl", "ru",
	"tel", "tv", "uk", "us",
}

// EOF
