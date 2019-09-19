package reqstatus

import (
	//"fmt"
	//"time"
	"strings"
	"strconv"
	//"net/http"
	//"io/ioutil"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/helper"
	"github.com/elastic/beats/libbeat/common"
	//"github.com/elastic/beats/libbeat/common/cfgwarn"
	"github.com/elastic/beats/metricbeat/mb"
	"github.com/elastic/beats/metricbeat/mb/parse"
)

const (
	// defaultScheme is the default scheme to use when it is not specified in
	// the host config.
	defaultScheme = "http"

	// defaultPath is the default path to the ngx_http_stub_status_module endpoint on Nginx.
	defaultPath = "/server-status"
)

var (
	hostParser = parse.URLHostParserBuilder{
		DefaultScheme: defaultScheme,
		PathConfigKey: "req_status_path",
		DefaultPath:   defaultPath,
	}.Build()
)

var logger = logp.NewLogger("tengine.reqstatus")

// init registers the MetricSet with the central registry as soon as the program
// starts. The New function will be called later to instantiate an instance of
// the MetricSet for each host defined in the module's configuration. After the
// MetricSet has been created then Fetch will begin to be called periodically.
func init() {
	mb.Registry.MustAddMetricSet("tengine", "reqstatus", New,
		mb.WithHostParser(hostParser),
		mb.DefaultMetricSet(),
	)
}

// MetricSet holds any configuration or state information. It must implement
// the mb.MetricSet interface. And this is best achieved by embedding
// mb.BaseMetricSet because it implements all of the required mb.MetricSet
// interface methods except for Fetch.
type MetricSet struct {
	mb.BaseMetricSet
	config ReqStatusConfig	
	http   *helper.HTTP
}

type ReqStatusConfig struct {
	Fields string `config:"req_status_fields"`
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := ReqStatusConfig {}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}
	http, err := helper.NewHTTP(base)
	if err != nil {
		return nil, err
	}

	return &MetricSet{
		BaseMetricSet: base,
		config:        config,
		http:          http,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch(report mb.ReporterV2) {
	logp.Info("modules", m.config.Fields)
	fields := strings.Split(m.config.Fields, " ")
	scanner, err := m.http.FetchScanner()
	if err != nil {
		return
	}
	//scanner.Scan()
	//resp := scanner.Text()
	//logp.Debug("modules", resp)
	//lines := strings.Split(resp, "\n")
	//logp.Info("modules", len(lines))
	//for _, line := range lines {
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			values := strings.Split(line, ",")
			tengineFields := make(common.MapStr)
			for i, field := range fields {
				if i == 0 {
					tengineFields[field] = values[i]
				} else {
					value, err := strconv.Atoi(values[i])
					if err != nil {
						tengineFields[field] = 0
					} else {
						tengineFields[field] = value
					}
				}
			}
			event := mb.Event{MetricSetFields: tengineFields}
			report.Event(event)
		}
	}
}
