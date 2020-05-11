package log

import "fmt"

//DefaultEncoder returns a default encoder for normal terminal/console log output
func DefaultEncoder() IColumnEncoder {
	ce := NewColumnEncoder().
		With(Column("time", TimeText("2006-01-02 15:04:05.000"))).
		With(Column("level", LevelText(5))).
		With(Column("logger", NameText(10))).
		With(Column("module", ModuleText(15))).
		With(Column("code", CodeText(30))).
		With(Column("message", MessageText(0)))
	return ce
}

//NewColumnEncoder ...
func NewColumnEncoder() IColumnEncoder {
	return columnEncoder{
		columns: []IColumn{},
	}
}

//TimeText writes the timestamp with specified format
func TimeText(fmt string) ITextValue {
	return timeText{fmt: fmt}
}

//LevelText writes the level of the log record
func LevelText(width int) ITextValue {
	return levelText{width: width}
}

//NameText write the name of the logger
func NameText(width int) ITextValue {
	return nameText{width: width}
}

//ModuleText writes the package and function name
func ModuleText(width int) ITextValue {
	return moduleText{width: width}
}

//CodeText writes the file name and line number
func CodeText(width int) ITextValue {
	return codeText{width: width}
}

//MessageText writes the log message
func MessageText(width int) ITextValue {
	return messageText{width: width}
}

//DataText writes the named log data value
func DataText(fmt, name string, width int) ITextValue {
	if fmt == "" {
		fmt = "%v"
	}
	return dataText{fmt: fmt, name: name, width: width}
}

//IColumnEncoder manages an array of encoders to make up one line of console logging
type IColumnEncoder interface {
	IEncoder
	Columns() []IColumn
	With(...IColumn) IColumnEncoder
}

//columnEncoder implements IColumnEncoder
type columnEncoder struct {
	columns []IColumn
}

func (ce columnEncoder) Columns() []IColumn { return ce.columns }

func (ce columnEncoder) With(cols ...IColumn) IColumnEncoder {
	ce.columns = append(ce.columns, cols...)
	return ce
}

//Column to add to list of columns
func Column(name string, text ITextValue) IColumn {
	return column{
		name: name,
		text: text,
	}
}

//Encode ...
func (ce columnEncoder) Encode(l ILogger, r Record) []byte {
	text := ""
	//multiple columns
	for _, col := range ce.columns {
		text += "|" + col.Text(l, r)
	}
	text += "\n"
	return []byte(text[1:])
}

//IColumn in a IColumnEncoder
type IColumn interface {
	ITextValue
	Name() string
}

//column implements IColumn
type column struct {
	name string
	text ITextValue
}

func (c column) Name() string {
	return c.name
}

func (c column) Text(l ILogger, r Record) string {
	return c.text.Text(l, r)
}

//ITextValue is used for each column
type ITextValue interface {
	Text(l ILogger, r Record) string
}

//============================================================================
type timeText struct {
	fmt string
}

func (c timeText) Text(l ILogger, r Record) string {
	if c.fmt == "" {
		return r.Time.Format("2006-01-02 15:04:05.000")
	}

	return r.Time.Format(c.fmt)
}

//============================================================================
type levelText struct {
	width int
}

func (c levelText) Text(l ILogger, r Record) string {
	return textField(c.width, fmt.Sprintf("%s", r.Level))
}

//============================================================================
type nameText struct {
	width int
}

func (c nameText) Text(l ILogger, r Record) string {
	return textField(c.width, l.Name())
}

//============================================================================
type moduleText struct {
	width int
}

func (c moduleText) Text(l ILogger, r Record) string {
	return textField(c.width, fmt.Sprintf("%s", r.Caller.Package+"."+r.Caller.Function+"()"))
}

//============================================================================
type codeText struct {
	width int
}

func (c codeText) Text(l ILogger, r Record) string {
	return textField(c.width, fmt.Sprintf("%s(%5d)", r.Caller.File, r.Caller.Line))
}

//============================================================================
type messageText struct {
	width int
}

func (c messageText) Text(l ILogger, r Record) string {
	return textField(c.width, r.Message)
}

//============================================================================
type dataText struct {
	fmt   string
	name  string
	width int
}

func (c dataText) Text(l ILogger, r Record) string {
	s := ""
	v, ok := l.Get(c.name)
	if ok {
		s = fmt.Sprintf(c.fmt, v)
	}
	return textField(c.width, s)
}

//============================================================================
func textField(w int, s string) string {
	if w <= 0 {
		return s
	}
	l := len(s)
	if l > w {
		s = s[l-w:]
	}
	return fmt.Sprintf("%-*.*s", w, w, s)
}
