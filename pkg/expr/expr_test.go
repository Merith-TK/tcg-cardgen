package expr_test

import (
	"testing"

	"github.com/Merith-TK/tcg-cardgen/pkg/expr"
)

func vars(kv ...string) map[string]string {
	m := make(map[string]string, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return m
}

// ─── Substitute ───────────────────────────────────────────────────────────────

func TestSubstituteSimple(t *testing.T) {
	p := expr.New(vars("card.title", "Lightning Bolt"))
	got := p.Substitute("Name: {{card.title}}")
	if got != "Name: Lightning Bolt" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteMissing(t *testing.T) {
	p := expr.New(vars())
	got := p.Substitute("{{unknown}}")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSubstituteFallback(t *testing.T) {
	p := expr.New(vars())
	got := p.Substitute("{{color|colorless}}")
	if got != "colorless" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteFallbackUsesValue(t *testing.T) {
	p := expr.New(vars("color", "red"))
	got := p.Substitute("{{color|colorless}}")
	if got != "red" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteTernaryTrue(t *testing.T) {
	p := expr.New(vars("mtg.color", "white"))
	got := p.Substitute("{{mtg.color == 'white' ? '#000000' : '#ffffff'}}")
	if got != "#000000" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteTernaryFalse(t *testing.T) {
	p := expr.New(vars("mtg.color", "blue"))
	got := p.Substitute("{{mtg.color == 'white' ? '#000000' : '#ffffff'}}")
	if got != "#ffffff" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteTernaryNotEqual(t *testing.T) {
	p := expr.New(vars("rarity", "rare"))
	got := p.Substitute("{{rarity != 'common' ? 'special' : 'normal'}}")
	if got != "special" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteTernaryTruthyCheck(t *testing.T) {
	p := expr.New(vars("mtg.power", "3"))
	got := p.Substitute("{{mtg.power ? mtg.power : '—'}}")
	if got != "3" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteTernaryFalsyCheck(t *testing.T) {
	p := expr.New(vars()) // mtg.power not set
	got := p.Substitute("{{mtg.power ? mtg.power : '—'}}")
	if got != "—" {
		t.Errorf("got %q", got)
	}
}

func TestSubstitutePathInString(t *testing.T) {
	p := expr.New(vars("mtg.color", "red"))
	got := p.Substitute("{{mtg.color|colorless}}_frame.png")
	if got != "red_frame.png" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteMultipleBlocks(t *testing.T) {
	p := expr.New(vars("a", "hello", "b", "world"))
	got := p.Substitute("{{a}} {{b}}!")
	if got != "hello world!" {
		t.Errorf("got %q", got)
	}
}

// ─── EvalBool ─────────────────────────────────────────────────────────────────

func TestEvalBoolSimpleTruthy(t *testing.T) {
	p := expr.New(vars("mtg.power", "3"))
	if !p.EvalBool("{{mtg.power}}") {
		t.Error("expected true")
	}
}

func TestEvalBoolSimpleFalsy(t *testing.T) {
	p := expr.New(vars())
	if p.EvalBool("{{mtg.power}}") {
		t.Error("expected false")
	}
}

func TestEvalBoolAND(t *testing.T) {
	p := expr.New(vars("a", "1", "b", "2"))
	if !p.EvalBool("{{a}} && {{b}}") {
		t.Error("expected true for a && b both set")
	}
}

func TestEvalBoolANDOneMissing(t *testing.T) {
	p := expr.New(vars("a", "1"))
	if p.EvalBool("{{a}} && {{b}}") {
		t.Error("expected false: b is missing")
	}
}

func TestEvalBoolOR(t *testing.T) {
	p := expr.New(vars("a", "1"))
	if !p.EvalBool("{{a}} || {{b}}") {
		t.Error("expected true: a is set")
	}
}

func TestEvalBoolORBothMissing(t *testing.T) {
	p := expr.New(vars())
	if p.EvalBool("{{a}} || {{b}}") {
		t.Error("expected false: neither set")
	}
}

func TestEvalBoolEquality(t *testing.T) {
	p := expr.New(vars("color", "red"))
	if !p.EvalBool("{{color == 'red'}}") {
		t.Error("expected true")
	}
}

func TestEvalBoolEqualityFalse(t *testing.T) {
	p := expr.New(vars("color", "blue"))
	if p.EvalBool("{{color == 'red'}}") {
		t.Error("expected false")
	}
}

func TestEvalBoolFalseLiteral(t *testing.T) {
	p := expr.New(vars("enabled", "false"))
	if p.EvalBool("{{enabled}}") {
		t.Error("expected false for literal 'false'")
	}
}
