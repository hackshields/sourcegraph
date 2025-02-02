<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# monitoring

```go
import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"
```

Package monitoring declares types for Sourcegraph's monitoring generator as well as the generator implementation itself\.

To learn more about developing monitoring\, see the guide: https://about.sourcegraph.com/handbook/engineering/observability/monitoring

To learn more about the generator\, see the top\-level program: https://github.com/sourcegraph/sourcegraph/tree/main/monitoring

## Index

- [func Generate(logger log15.Logger, opts GenerateOptions, containers ...*Container) error](<#func-generate>)
- [func Int64Ptr(i int64) *int64](<#func-int64ptr>)
- [func StringPtr(s string) *string](<#func-stringptr>)
- [type Container](<#type-container>)
- [type GenerateOptions](<#type-generateoptions>)
- [type Group](<#type-group>)
- [type Observable](<#type-observable>)
- [type ObservableAlertDefinition](<#type-observablealertdefinition>)
  - [func Alert() *ObservableAlertDefinition](<#func-alert>)
  - [func (a *ObservableAlertDefinition) For(d time.Duration) *ObservableAlertDefinition](<#func-observablealertdefinition-for>)
  - [func (a *ObservableAlertDefinition) Greater(f float64, aggregator *string) *ObservableAlertDefinition](<#func-observablealertdefinition-greater>)
  - [func (a *ObservableAlertDefinition) GreaterOrEqual(f float64, aggregator *string) *ObservableAlertDefinition](<#func-observablealertdefinition-greaterorequal>)
  - [func (a *ObservableAlertDefinition) Less(f float64, aggregator *string) *ObservableAlertDefinition](<#func-observablealertdefinition-less>)
  - [func (a *ObservableAlertDefinition) LessOrEqual(f float64, aggregator *string) *ObservableAlertDefinition](<#func-observablealertdefinition-lessorequal>)
- [type ObservableOwner](<#type-observableowner>)
- [type ObservablePanel](<#type-observablepanel>)
  - [func Panel() ObservablePanel](<#func-panel>)
  - [func PanelMinimal() ObservablePanel](<#func-panelminimal>)
  - [func (p ObservablePanel) Interval(ms int) ObservablePanel](<#func-observablepanel-interval>)
  - [func (p ObservablePanel) LegendFormat(format string) ObservablePanel](<#func-observablepanel-legendformat>)
  - [func (p ObservablePanel) Max(max float64) ObservablePanel](<#func-observablepanel-max>)
  - [func (p ObservablePanel) Min(min float64) ObservablePanel](<#func-observablepanel-min>)
  - [func (p ObservablePanel) MinAuto() ObservablePanel](<#func-observablepanel-minauto>)
  - [func (p ObservablePanel) Unit(t UnitType) ObservablePanel](<#func-observablepanel-unit>)
  - [func (p ObservablePanel) With(ops ...ObservablePanelOption) ObservablePanel](<#func-observablepanel-with>)
- [type ObservablePanelOption](<#type-observablepaneloption>)
- [type Row](<#type-row>)
- [type UnitType](<#type-unittype>)


## func [Generate](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/generator.go#L40>)

```go
func Generate(logger log15.Logger, opts GenerateOptions, containers ...*Container) error
```

Generate is the main Sourcegraph monitoring generator entrypoint\.

## func [Int64Ptr](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/util.go#L29>)

```go
func Int64Ptr(i int64) *int64
```

IntPtr converts an int64 value to a pointer\, useful for setting fields in some APIs\.

## func [StringPtr](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/util.go#L23>)

```go
func StringPtr(s string) *string
```

StringPtr converts a string value to a pointer\, useful for setting fields in some APIs\.

## type [Container](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L17-L43>)

Container describes a Docker container to be observed\.

These correspond to dashboards in Grafana\.

```go
type Container struct {
    // Name of the Docker container, e.g. "syntect-server".
    Name string

    // Title of the Docker container, e.g. "Syntect Server".
    Title string

    // Description of the Docker container. It should describe what the container
    // is responsible for, so that the impact of issues in it is clear.
    Description string

    // List of Annotations to apply to the dashboard.
    Annotations []sdk.Annotation

    // List of Template Variables to apply to the dashboard
    Templates []sdk.TemplateVar

    // Groups of observable information about the container.
    Groups []Group

    // NoSourcegraphDebugServer indicates if this container does not export the standard
    // Sourcegraph debug server (package `internal/debugserver`).
    //
    // This is used to configure monitoring features that depend on information exported
    // by the standard Sourcegraph debug server.
    NoSourcegraphDebugServer bool
}
```

## type [GenerateOptions](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/generator.go#L25-L37>)

GenerateOptions declares options for the monitoring generator\.

```go
type GenerateOptions struct {
    // Toggles pruning of dangling generated assets through simple heuristic, should be disabled during builds
    DisablePrune bool
    // Trigger reload of active Prometheus or Grafana instance (requires respective output directories)
    Reload bool

    // Output directory for generated Grafana assets
    GrafanaDir string
    // Output directory for generated Prometheus assets
    PrometheusDir string
    // Output directory for generated documentation
    DocsDir string
}
```

## type [Group](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L354-L369>)

Group describes a group of observable information about a container\.

These correspond to collapsible sections in a Grafana dashboard\.

```go
type Group struct {
    // Title of the group, briefly summarizing what this group is about, or
    // "General" if the group is just about the container in general.
    Title string

    // Hidden indicates whether or not the group should be hidden by default.
    //
    // This should only be used when the dashboard is already full of information
    // and the information presented in this group is unlikely to be the cause of
    // issues and should generally only be inspected in the event that an alert
    // for that information is firing.
    Hidden bool

    // Rows of observable metrics.
    Rows []Row
}
```

## type [Observable](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L434-L549>)

Observable describes a metric about a container that can be observed\. For example\, memory usage\.

These correspond to Grafana graphs\.

```go
type Observable struct {
    // Name is a short and human-readable lower_snake_case name describing what is being observed.
    //
    // It must be unique relative to the service name.
    //
    // Good examples:
    //
    //  github_rate_limit_remaining
    // 	search_error_rate
    //
    // Bad examples:
    //
    //  repo_updater_github_rate_limit
    // 	search_error_rate_over_5m
    //
    Name string

    // Description is a human-readable description of exactly what is being observed.
    // If a query groups by a label (such as with a `sum by(...)`), ensure that this is
    // reflected in the description by noting that this observable is grouped "by ...".
    //
    // Good examples:
    //
    // 	"remaining GitHub API rate limit quota"
    // 	"number of search errors every 5m"
    //  "90th percentile search request duration over 5m"
    //  "internal API error responses every 5m by route"
    //
    // Bad examples:
    //
    // 	"GitHub rate limit"
    // 	"search errors[5m]"
    // 	"P90 search latency"
    //
    Description string

    // Owner indicates the team that owns this Observable (including its alerts and maintainence).
    Owner ObservableOwner

    // Query is the actual Prometheus query that should be observed.
    Query string

    // DataMustExist indicates if the query must return data.
    //
    // For example, repo_updater_memory_usage should always have data present and an alert should
    // fire if for some reason that query is not returning any data, so this would be set to true.
    // In contrast, search_error_rate would depend on users actually performing searches and we
    // would not want an alert to fire if no data was present, so this will not need to be set.
    DataMustExist bool

    // Warning and Critical alert definitions.
    // Consider adding at least a Warning or Critical alert to each Observable to make it
    // easy to identify when the target of this metric is misbehaving. If no alerts are
    // provided, NoAlert must be set and Interpretation must be provided.
    Warning, Critical *ObservableAlertDefinition

    // NoAlerts must be set by Observables that do not have any alerts.
    // This ensures the omission of alerts is intentional. If set to true, an Interpretation
    // must be provided in place of PossibleSolutions.
    NoAlert bool

    // PossibleSolutions is Markdown describing possible solutions in the event that the
    // alert is firing. This field not required if no alerts are attached to this Observable.
    // If there is no clear potential resolution or there is no alert configured, "none"
    // must be explicitly stated.
    //
    // Use the Interpretation field for additional guidance on understanding this Observable
    // that isn't directly related to solving it.
    //
    // Contacting support should not be mentioned as part of a possible solution, as it is
    // communicated elsewhere.
    //
    // To make writing the Markdown more friendly in Go, string literals like this:
    //
    // 	Observable{
    // 		PossibleSolutions: `
    // 			- Foobar 'some code'
    // 		`
    // 	}
    //
    // Becomes:
    //
    // 	- Foobar `some code`
    //
    // In other words:
    //
    // 1. The preceding newline is removed.
    // 2. The indentation in the string literal is removed (based on the last line).
    // 3. Single quotes become backticks.
    // 4. The last line (which is all indention) is removed.
    // 5. Non-list items are converted to a list.
    //
    PossibleSolutions string

    // Interpretation is Markdown that can serve as a reference for interpreting this
    // observable. For example, Interpretation could provide guidance on what sort of
    // patterns to look for in the observable's graph and document why this observable is
    // usefule.
    //
    // If no alerts are configured for an observable, this field is required. If the
    // Description is sufficient to capture what this Observable describes, "none" must be
    // explicitly stated.
    //
    // To make writing the Markdown more friendly in Go, string literal processing as
    // PossibleSolutions is provided, though the output is not converted to a list.
    Interpretation string

    // Panel provides options for how to render the metric in the Grafana panel.
    // A recommended set of options and customizations are available from the `Panel()`
    // constructor.
    //
    // Additional customizations can be made via `ObservablePanel.With()` for cases where
    // the provided `ObservablePanel` is insufficient - see `ObservablePanelOption` for
    // more details.
    Panel ObservablePanel
}
```

## type [ObservableAlertDefinition](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L613-L628>)

ObservableAlertDefinition defines when an alert would be considered firing\.

```go
type ObservableAlertDefinition struct {
    // contains filtered or unexported fields
}
```

### func [Alert](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L608>)

```go
func Alert() *ObservableAlertDefinition
```

Alert provides a builder for defining alerting on an Observable\.

### func \(\*ObservableAlertDefinition\) [For](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L684>)

```go
func (a *ObservableAlertDefinition) For(d time.Duration) *ObservableAlertDefinition
```

For indicates how long the given thresholds must be exceeded for this alert to be considered firing\. Defaults to 0s \(immediately alerts when threshold is exceeded\)\.

### func \(\*ObservableAlertDefinition\) [Greater](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L657>)

```go
func (a *ObservableAlertDefinition) Greater(f float64, aggregator *string) *ObservableAlertDefinition
```

Greater indicates the alert should fire when strictly greater to this value\.

### func \(\*ObservableAlertDefinition\) [GreaterOrEqual](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L631>)

```go
func (a *ObservableAlertDefinition) GreaterOrEqual(f float64, aggregator *string) *ObservableAlertDefinition
```

GreaterOrEqual indicates the alert should fire when greater or equal the given value\.

### func \(\*ObservableAlertDefinition\) [Less](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L670>)

```go
func (a *ObservableAlertDefinition) Less(f float64, aggregator *string) *ObservableAlertDefinition
```

Less indicates the alert should fire when strictly less than this value\.

### func \(\*ObservableAlertDefinition\) [LessOrEqual](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L644>)

```go
func (a *ObservableAlertDefinition) LessOrEqual(f float64, aggregator *string) *ObservableAlertDefinition
```

LessOrEqual indicates the alert should fire when less than or equal to the given value\.

## type [ObservableOwner](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L402>)

ObservableOwner denotes a team that owns an Observable\. The current teams are described in the handbook: https://about.sourcegraph.com/company/team/org_chart#engineering

```go
type ObservableOwner string
```

```go
const (
    ObservableOwnerSearch          ObservableOwner = "search"
    ObservableOwnerBatches         ObservableOwner = "batches"
    ObservableOwnerCodeIntel       ObservableOwner = "code-intel"
    ObservableOwnerDistribution    ObservableOwner = "distribution"
    ObservableOwnerSecurity        ObservableOwner = "security"
    ObservableOwnerWeb             ObservableOwner = "web"
    ObservableOwnerCoreApplication ObservableOwner = "core application"
)
```

## type [ObservablePanel](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L12-L17>)

ObservablePanel declares options for visualizing an Observable\, as well as some default customization options\. A default panel can be instantiated with the \`Panel\(\)\` constructor\, and further customized using \`ObservablePanel\.With\(ObservablePanelOption\)\`\.

```go
type ObservablePanel struct {
    // contains filtered or unexported fields
}
```

### func [Panel](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L21>)

```go
func Panel() ObservablePanel
```

Panel provides a builder for customizing an Observable visualization\, starting with recommended defaults\.

### func [PanelMinimal](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L35>)

```go
func PanelMinimal() ObservablePanel
```

PanelMinimal provides a builder for customizing an Observable visualization starting with an extremely minimal graph panel\.

In general\, we advise using Panel\(\) instead to start with recommended defaults\.

### func \(ObservablePanel\) [Interval](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L89>)

```go
func (p ObservablePanel) Interval(ms int) ObservablePanel
```

Interval declares the panel's interval in milliseconds\.

### func \(ObservablePanel\) [LegendFormat](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L72>)

```go
func (p ObservablePanel) LegendFormat(format string) ObservablePanel
```

LegendFormat sets the panel's legend format\, which may use Go template strings to select labels from the Prometheus query\.

### func \(ObservablePanel\) [Max](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L63>)

```go
func (p ObservablePanel) Max(max float64) ObservablePanel
```

Max sets the maximum value of the Y axis on the panel\. The default is auto\.

### func \(ObservablePanel\) [Min](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L44>)

```go
func (p ObservablePanel) Min(min float64) ObservablePanel
```

Min sets the minimum value of the Y axis on the panel\. The default is zero\.

### func \(ObservablePanel\) [MinAuto](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L55>)

```go
func (p ObservablePanel) MinAuto() ObservablePanel
```

Min sets the minimum value of the Y axis on the panel to auto\, instead of the default zero\.

This is generally only useful if trying to show negative numbers\.

### func \(ObservablePanel\) [Unit](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L80>)

```go
func (p ObservablePanel) Unit(t UnitType) ObservablePanel
```

Unit sets the panel's Y axis unit type\.

### func \(ObservablePanel\) [With](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel.go#L104>)

```go
func (p ObservablePanel) With(ops ...ObservablePanelOption) ObservablePanel
```

With adds the provided options to be applied when building this panel\.

Before using this\, check if the customization you want is already included in the default \`Panel\(\)\` or available as a function on \`ObservablePanel\`\, such as \`ObservablePanel\.Unit\(UnitType\)\` for setting the units on a panel\.

Shared customizations are exported by \`PanelOptions\`\, or you can write your option \- see \`ObservablePanelOption\` documentation for more details\.

## type [ObservablePanelOption](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/panel_options.go#L34>)

ObservablePanelOption declares an option for customizing a graph panel\. \`ObservablePanel\` is responsible for collecting and applying options\.

You can make any customization you want to a graph panel by using \`ObservablePanel\.With\`:

```
Panel: monitoring.Panel().With(func(o monitoring.Observable, g *sdk.GraphPanel) {
  // modify 'g' with desired changes
}),
```

When writing a custom \`ObservablePanelOption\`\, keep in mind that:

\- There are only ever two \`YAxes\`: left at \`YAxes\[0\]\` and right at \`YAxes\[1\]\`\. Target customizations at the Y\-axis you want to modify\, e\.g\. \`YAxes\[0\]\.Property = Value\`\.

\- The observable being graphed is configured in \`Targets\[0\]\`\. Customize it by editing it directly\, e\.g\. \`Targets\[0\]\.Property = Value\`\.

If an option could be leveraged by multiple observables\, a shared panel option can be defined in the \`monitoring\` package\.

When creating a shared \`ObservablePanelOption\`\, it should defined as a function on the \`panelOptionsLibrary\` that returns a \`ObservablePanelOption\`\. The function should be It can then be used with the \`ObservablePanel\.With\`:

```
Panel: monitoring.Panel().With(monitoring.PanelOptions.MyCustomization),
```

Using a shared prefix helps with discoverability of available options\.

```go
type ObservablePanelOption func(Observable, *sdk.GraphPanel)
```

## type [Row](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/monitoring.go#L386>)

Row of observable metrics\.

These correspond to a row of Grafana graphs\.

```go
type Row []Observable
```

## type [UnitType](<https://github.com/sourcegraph/sourcegraph/blob/main/monitoring/monitoring/dashboards.go#L11>)

UnitType for controlling the unit type display on graphs\.

```go
type UnitType string
```

From https://sourcegraph.com/github.com/grafana/grafana@b63b82976b3708b082326c0b7d42f38d4bc261fa/-/blob/packages/grafana-data/src/valueFormats/categories.ts#L23

```go
const (
    // Number is the default unit type.
    Number UnitType = "short"

    // Milliseconds for representing time.
    Milliseconds UnitType = "ms"

    // Seconds for representing time.
    Seconds UnitType = "s"

    // Percentage in the range of 0-100.
    Percentage UnitType = "percent"

    // Bytes in IEC (1024) format, e.g. for representing storage sizes.
    Bytes UnitType = "bytes"

    // BitsPerSecond, e.g. for representing network and disk IO.
    BitsPerSecond UnitType = "bps"

    // ReadsPerSecond, e.g for representing disk IO.
    ReadsPerSecond = "rps"

    // WritesPerSecond, e.g for representing disk IO.
    WritesPerSecond = "wps"
)
```



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
