package pattern

import (
	"testing"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/models"
	"github.com/denis-tingajkin/go-header/text"
)

func holderSampleConfig(holders ...string) models.ReadOnlyConfiguration {
	config := models.Configuration{
		CopyrigtHolders: holders,
	}
	return models.AsReadonly(&config)
}

func TestHolder1(t *testing.T) {
	rule := CopyrightHolder(holderSampleConfig())
	reader := text.NewReader("any copyright holder")
	errs := rule.Verify(reader)
	if !errs.Empty() {
		t.FailNow()
	}
}
func TestHolder2(t *testing.T) {
	conf := holderSampleConfig("1", "2", "3")
	rule := CopyrightHolder(conf)
	reader := text.NewReader("1\n2\n3\n4\n")
	for i := 1; i < 4; i++ {
		errs := rule.Verify(reader)
		if !errs.Empty() {
			println(errs.String())
			t.FailNow()
		}
		_ = reader.Next()
	}
	p := reader.Position()
	errs := rule.Verify(reader)
	if errs.Empty() {
		t.FailNow()
	}
	if errs.String() != messages.UnknownCopyrightHolder(p, "4", conf.CopyrightHolders()...).Error() {
		t.FailNow()
	}
}
func TestHolder3(t *testing.T) {
	rule := CopyrightHolder(holderSampleConfig())
	reader := text.NewReader("")
	errs := rule.Verify(reader)
	if errs.Empty() {
		t.FailNow()
	}
	if messages.UnknownCopyrightHolder(0, "").Error() != errs.String() {
		t.FailNow()
	}
}

func TestHolder4(t *testing.T) {
	conf := holderSampleConfig("1", "2", "3")
	rule := CopyrightHolder(conf)
	reader := text.NewReader("1\n2\n3\n1")
	for i := 1; i < 4; i++ {
		errs := rule.Verify(reader)
		if !errs.Empty() {
			println(errs.String())
			t.FailNow()
		}
		_ = reader.Next()
	}
	errs := rule.Verify(reader)
	if errs.Empty() {
		t.FailNow()
	}
	if errs.String() != messages.NewErrorList(messages.CopyrightHolderAlreadyInUse("1")).String() {
		t.FailNow()
	}
}
