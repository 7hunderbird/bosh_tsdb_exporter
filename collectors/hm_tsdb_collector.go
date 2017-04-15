package collectors

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type HMMetric struct {
	Name       string
	Value      float64
	Deployment string
	Job        string
	Index      string
	Id         string
}

type HMTSDBCollector struct {
	tsdbListener                             net.Listener
	jobHealthyMetric                         *prometheus.GaugeVec
	jobLoadAvg01Metric                       *prometheus.GaugeVec
	jobCPUSysMetric                          *prometheus.GaugeVec
	jobCPUUserMetric                         *prometheus.GaugeVec
	jobCPUWaitMetric                         *prometheus.GaugeVec
	jobMemKBMetric                           *prometheus.GaugeVec
	jobMemPercentMetric                      *prometheus.GaugeVec
	jobSwapKBMetric                          *prometheus.GaugeVec
	jobSwapPercentMetric                     *prometheus.GaugeVec
	jobSystemDiskInodePercentMetric          *prometheus.GaugeVec
	jobSystemDiskPercentMetric               *prometheus.GaugeVec
	jobEphemeralDiskInodePercentMetric       *prometheus.GaugeVec
	jobEphemeralDiskPercentMetric            *prometheus.GaugeVec
	jobPersistentDiskInodePercentMetric      *prometheus.GaugeVec
	jobPersistentDiskPercentMetric           *prometheus.GaugeVec
	totalHMTSDBReceivedMessagesMetric        prometheus.Counter
	totalHMTSDBInvalidMessagesMetric         prometheus.Counter
	totalHMTSDBDiscardedMessagesMetric       prometheus.Counter
	lastHMTSDBReceivedMessageTimestampMetric prometheus.Gauge
	lastHMTSDBScrapeTimestampMetric          prometheus.Gauge
	lastHMTSDBScrapeDurationSecondsMetric    prometheus.Gauge
}

func NewHMTSDBCollector(
	namespace string,
	environment string,
	tsdbListener net.Listener,
) *HMTSDBCollector {
	jobHealthyMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "healthy",
			Help:      "BOSH Job Healthy (1 for healthy, 0 for unhealthy).",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobLoadAvg01Metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "load_avg01",
			Help:      "BOSH Job Load avg01.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobCPUSysMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "cpu_sys",
			Help:      "BOSH Job CPU System.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobCPUUserMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "cpu_user",
			Help:      "BOSH Job CPU User.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobCPUWaitMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "cpu_wait",
			Help:      "BOSH Job CPU Wait.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobMemKBMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "mem_kb",
			Help:      "BOSH Job Memory KB.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobMemPercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "mem_percent",
			Help:      "BOSH Job Memory Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobSwapKBMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "swap_kb",
			Help:      "BOSH Job Swap KB.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobSwapPercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "swap_percent",
			Help:      "BOSH Job Swap Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobSystemDiskInodePercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "system_disk_inode_percent",
			Help:      "BOSH Job System Disk Inode Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobSystemDiskPercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "system_disk_percent",
			Help:      "BOSH Job System Disk Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobEphemeralDiskInodePercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "ephemeral_disk_inode_percent",
			Help:      "BOSH Job Ephemeral Disk Inode Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobEphemeralDiskPercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "ephemeral_disk_percent",
			Help:      "BOSH Job Ephemeral Disk Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobPersistentDiskInodePercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "persistent_disk_inode_percent",
			Help:      "BOSH Job Persistent Disk Inode Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	jobPersistentDiskPercentMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "job",
			Name:      "persistent_disk_percent",
			Help:      "BOSH Job Persistent Disk Percent.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
		[]string{"bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_index"},
	)

	totalHMTSDBReceivedMessagesMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "hm_tsdb",
			Name:      "received_messages_total",
			Help:      "Total number of BOSH HM TSDB received messages.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
	)

	totalHMTSDBInvalidMessagesMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "hm_tsdb",
			Name:      "invalid_messages_total",
			Help:      "Total number of BOSH HM TSDB invalid messages.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
	)

	totalHMTSDBDiscardedMessagesMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "hm_tsdb",
			Name:      "discarded_messages_total",
			Help:      "Total number of BOSH HM TSDB discarded messages.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
	)

	lastHMTSDBReceivedMessageTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "hm_tsdb",
			Name:      "last_received_message_timestamp",
			Help:      "Number of seconds since 1970 since last received message from BOSH HM TSDB.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
	)

	lastHMTSDBScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_hm_tsdb_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of BOSH HM TSDB collector.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
	)

	lastHMTSDBScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_hm_tsdb_scrape_duration_seconds",
			Help:      "Duration of the last scrape of BOSH HM TSDB collector.",
			ConstLabels: prometheus.Labels{
				"environment": environment,
			},
		},
	)

	collector := &HMTSDBCollector{
		tsdbListener:                             tsdbListener,
		jobHealthyMetric:                         jobHealthyMetric,
		jobLoadAvg01Metric:                       jobLoadAvg01Metric,
		jobCPUSysMetric:                          jobCPUSysMetric,
		jobCPUUserMetric:                         jobCPUUserMetric,
		jobCPUWaitMetric:                         jobCPUWaitMetric,
		jobMemKBMetric:                           jobMemKBMetric,
		jobMemPercentMetric:                      jobMemPercentMetric,
		jobSwapKBMetric:                          jobSwapKBMetric,
		jobSwapPercentMetric:                     jobSwapPercentMetric,
		jobSystemDiskInodePercentMetric:          jobSystemDiskInodePercentMetric,
		jobSystemDiskPercentMetric:               jobSystemDiskPercentMetric,
		jobEphemeralDiskInodePercentMetric:       jobEphemeralDiskInodePercentMetric,
		jobEphemeralDiskPercentMetric:            jobEphemeralDiskPercentMetric,
		jobPersistentDiskInodePercentMetric:      jobPersistentDiskInodePercentMetric,
		jobPersistentDiskPercentMetric:           jobPersistentDiskPercentMetric,
		totalHMTSDBReceivedMessagesMetric:        totalHMTSDBReceivedMessagesMetric,
		totalHMTSDBInvalidMessagesMetric:         totalHMTSDBInvalidMessagesMetric,
		totalHMTSDBDiscardedMessagesMetric:       totalHMTSDBDiscardedMessagesMetric,
		lastHMTSDBReceivedMessageTimestampMetric: lastHMTSDBReceivedMessageTimestampMetric,
		lastHMTSDBScrapeTimestampMetric:          lastHMTSDBScrapeTimestampMetric,
		lastHMTSDBScrapeDurationSecondsMetric:    lastHMTSDBScrapeDurationSecondsMetric,
	}
	go collector.listenHMTSDB()

	return collector
}

func (c *HMTSDBCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	c.jobHealthyMetric.Collect(ch)
	c.jobLoadAvg01Metric.Collect(ch)
	c.jobCPUSysMetric.Collect(ch)
	c.jobCPUUserMetric.Collect(ch)
	c.jobCPUWaitMetric.Collect(ch)
	c.jobMemKBMetric.Collect(ch)
	c.jobMemPercentMetric.Collect(ch)
	c.jobSwapKBMetric.Collect(ch)
	c.jobSwapPercentMetric.Collect(ch)
	c.jobSystemDiskInodePercentMetric.Collect(ch)
	c.jobSystemDiskPercentMetric.Collect(ch)
	c.jobEphemeralDiskInodePercentMetric.Collect(ch)
	c.jobEphemeralDiskPercentMetric.Collect(ch)
	c.jobPersistentDiskInodePercentMetric.Collect(ch)
	c.jobPersistentDiskPercentMetric.Collect(ch)

	c.totalHMTSDBReceivedMessagesMetric.Collect(ch)
	c.totalHMTSDBInvalidMessagesMetric.Collect(ch)
	c.totalHMTSDBDiscardedMessagesMetric.Collect(ch)
	c.lastHMTSDBReceivedMessageTimestampMetric.Collect(ch)

	c.lastHMTSDBScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastHMTSDBScrapeTimestampMetric.Collect(ch)

	c.lastHMTSDBScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastHMTSDBScrapeDurationSecondsMetric.Collect(ch)

	c.jobHealthyMetric.Reset()
	c.jobLoadAvg01Metric.Reset()
	c.jobCPUSysMetric.Reset()
	c.jobCPUUserMetric.Reset()
	c.jobCPUWaitMetric.Reset()
	c.jobMemKBMetric.Reset()
	c.jobMemPercentMetric.Reset()
	c.jobSwapKBMetric.Reset()
	c.jobSwapPercentMetric.Reset()
	c.jobSystemDiskInodePercentMetric.Reset()
	c.jobSystemDiskPercentMetric.Reset()
	c.jobEphemeralDiskInodePercentMetric.Reset()
	c.jobEphemeralDiskPercentMetric.Reset()
	c.jobPersistentDiskInodePercentMetric.Reset()
	c.jobPersistentDiskPercentMetric.Reset()
}

func (c *HMTSDBCollector) Describe(ch chan<- *prometheus.Desc) {
	c.jobHealthyMetric.Describe(ch)
	c.jobLoadAvg01Metric.Describe(ch)
	c.jobCPUSysMetric.Describe(ch)
	c.jobCPUUserMetric.Describe(ch)
	c.jobCPUWaitMetric.Describe(ch)
	c.jobMemKBMetric.Describe(ch)
	c.jobMemPercentMetric.Describe(ch)
	c.jobSwapKBMetric.Describe(ch)
	c.jobSwapPercentMetric.Describe(ch)
	c.jobSystemDiskInodePercentMetric.Describe(ch)
	c.jobSystemDiskPercentMetric.Describe(ch)
	c.jobEphemeralDiskInodePercentMetric.Describe(ch)
	c.jobEphemeralDiskPercentMetric.Describe(ch)
	c.jobPersistentDiskInodePercentMetric.Describe(ch)
	c.jobPersistentDiskPercentMetric.Describe(ch)
	c.totalHMTSDBReceivedMessagesMetric.Describe(ch)
	c.totalHMTSDBInvalidMessagesMetric.Describe(ch)
	c.totalHMTSDBDiscardedMessagesMetric.Describe(ch)
	c.lastHMTSDBReceivedMessageTimestampMetric.Describe(ch)
	c.lastHMTSDBScrapeTimestampMetric.Describe(ch)
	c.lastHMTSDBScrapeDurationSecondsMetric.Describe(ch)
}

func (c *HMTSDBCollector) listenHMTSDB() {
	for {
		conn, err := c.tsdbListener.Accept()
		if err != nil {
			log.Errorf("Error accepting BOSH HM TSDB connections: %v", err)
			continue
		}
		go c.handleHMMessage(conn)
	}
}

func (c *HMTSDBCollector) handleHMMessage(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		c.totalHMTSDBReceivedMessagesMetric.Inc()
		c.lastHMTSDBReceivedMessageTimestampMetric.Set(float64(time.Now().Unix()))

		hmMessage := scanner.Text()
		hmMetric, err := c.parseHMMessage(hmMessage)
		if err != nil {
			log.Error(err)
			c.totalHMTSDBInvalidMessagesMetric.Inc()
			continue
		}

		switch hmMetric.Name {
		case "system.healthy":
			c.jobHealthyMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.load.1m":
			c.jobLoadAvg01Metric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.cpu.sys":
			c.jobCPUSysMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.cpu.user":
			c.jobCPUUserMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.cpu.wait":
			c.jobCPUWaitMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.mem.kb":
			c.jobMemKBMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.mem.percent":
			c.jobMemPercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.swap.kb":
			c.jobSwapKBMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.swap.percent":
			c.jobSwapPercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.disk.system.inode_percent":
			c.jobSystemDiskInodePercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.disk.system.percent":
			c.jobSystemDiskPercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.disk.ephemeral.inode_percent":
			c.jobEphemeralDiskInodePercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.disk.ephemeral.percent":
			c.jobEphemeralDiskPercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.disk.persistent.inode_percent":
			c.jobPersistentDiskInodePercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		case "system.disk.persistent.percent":
			c.jobPersistentDiskPercentMetric.WithLabelValues(
				hmMetric.Deployment,
				hmMetric.Job,
				hmMetric.Id,
				hmMetric.Index,
			).Set(hmMetric.Value)
		default:
			log.Errorf("BOSH HM TSDB metric `%s` not supported, discarded", hmMetric.Name)
			c.totalHMTSDBDiscardedMessagesMetric.Inc()
		}
	}
}

func (c *HMTSDBCollector) parseHMMessage(hmMessage string) (HMMetric, error) {
	hmMetric := HMMetric{}

	log.Debugf("Parsing BOSH HM TSDB message `%s`", hmMessage)

	tokens := strings.Split(hmMessage, " ")
	if len(tokens) < 4 {
		return hmMetric, errors.New(fmt.Sprintf("BOSH HM TSDB message discarded, it has less than 4 tokens: %v", hmMessage))
	}

	hmMetric.Name = tokens[1]

	value, err := strconv.ParseFloat(tokens[3], 64)
	if err != nil {
		return hmMetric, errors.New(fmt.Sprintf("BOSH HM TSDB message discarded, value `%s` cannot be parsed as float: %v", tokens[3], err))
	}
	hmMetric.Value = value

	for i := 4; i < len(tokens); i++ {
		tag := strings.Split(tokens[i], "=")
		if len(tag) > 1 {
			switch tag[0] {
			case "deployment":
				hmMetric.Deployment = tag[1]
			case "job":
				hmMetric.Job = tag[1]
			case "index":
				hmMetric.Index = tag[1]
			case "id":
				hmMetric.Id = tag[1]
			}
		}
	}

	return hmMetric, nil
}
