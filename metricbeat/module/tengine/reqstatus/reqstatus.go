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
	fields []string
	preMetricSet map[string]common.MapStr
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
	fields := strings.Split(m.config.Fields, " ")

	return &MetricSet{
		BaseMetricSet: base,
		config:        config,
		fields:        fields,
		preMetricSet   make(map[string]common.MapStr)
		http:          http,
	}, nil
}

func makeMetricSet(line string, m *MetricSet) (common.MapStr, error) {
	values := strings.Split(line, ",")
	kv := values[0]

	currentSet := make(common.MapStr)
	retSet := make(common.MapStr)
	for i, value := range values {
		if i == 0 {
			currentSet[m.fields[i]] = value
			continue
		}
		intValue, err := strconv.Atoi(value)
		if err != nil {
			currentSet[m.fields[i]] = nil
		} else {
			currentSet[m.fields[i]] = intValue
		}
	}

	preSet, ok = m.preMetricSet[kv]
	if ok {
		for k, v := range currentSet {
			if k == kv {
				retSet[k] = k
			} else {
				retSet[k] = v - preSet[k]
			}
		}
	}
	m.preMetricSet[kv] = currentSet
	
	if ok {
		return retSet, nil
	} else {
		return nil, fmt.Errorf("have not pre set")
	}
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch(report mb.ReporterV2) {
	scanner, err := m.http.FetchScanner()
	if err != nil {
		logp.Info("tengine_reqstatus", err)
		return
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			if retMetricSet, err := makeMetricSet(line, m); err == nil {
				event := mb.Event{MetricSetFields: retMetricSet}
				report.Event(event)
			}
		}
	}
}
