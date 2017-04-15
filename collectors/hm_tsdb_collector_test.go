package collectors_test

import (
	"flag"
	"fmt"
	"net"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus"

	. "github.com/cloudfoundry-community/bosh_tsdb_exporter/collectors"
	. "github.com/cloudfoundry-community/bosh_tsdb_exporter/utils/test_matchers"
)

func init() {
	flag.Set("log.level", "fatal")
}

var _ = Describe("HMTSDBCollector", func() {
	var (
		err             error
		namespace       string
		environment     string
		tsdbListener    net.Listener
		hmTSDBCollector *HMTSDBCollector

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

		deploymentName                = "fake-deployment-name"
		jobName                       = "fake-job-name"
		jobID                         = "fake-job-id"
		jobIndex                      = "0"
		jobLoadAvg01                  = float64(0.01)
		jobCPUSys                     = float64(0.5)
		jobCPUUser                    = float64(1.0)
		jobCPUWait                    = float64(1.5)
		jobMemKB                      = 1000
		jobMemPercent                 = 10
		jobSwapKB                     = 2000
		jobSwapPercent                = 20
		jobSystemDiskInodePercent     = 10
		jobSystemDiskPercent          = 20
		jobEphemeralDiskInodePercent  = 30
		jobEphemeralDiskPercent       = 40
		jobPersistentDiskInodePercent = 50
		jobPersistentDiskPercent      = 60
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		environment = "test_environment"
		tsdbListener, err = net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())

		jobHealthyMetric = prometheus.NewGaugeVec(
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

		jobHealthyMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(1))

		jobLoadAvg01Metric = prometheus.NewGaugeVec(
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

		jobLoadAvg01Metric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(jobLoadAvg01)

		jobCPUSysMetric = prometheus.NewGaugeVec(
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

		jobCPUSysMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(jobCPUSys)

		jobCPUUserMetric = prometheus.NewGaugeVec(
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

		jobCPUUserMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(jobCPUUser)

		jobCPUWaitMetric = prometheus.NewGaugeVec(
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

		jobCPUWaitMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(jobCPUWait)

		jobMemKBMetric = prometheus.NewGaugeVec(
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

		jobMemKBMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobMemKB))

		jobMemPercentMetric = prometheus.NewGaugeVec(
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

		jobMemPercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobMemPercent))

		jobSwapKBMetric = prometheus.NewGaugeVec(
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

		jobSwapKBMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobSwapKB))

		jobSwapPercentMetric = prometheus.NewGaugeVec(
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

		jobSwapPercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobSwapPercent))

		jobSystemDiskInodePercentMetric = prometheus.NewGaugeVec(
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

		jobSystemDiskInodePercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobSystemDiskInodePercent))

		jobSystemDiskPercentMetric = prometheus.NewGaugeVec(
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

		jobSystemDiskPercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobSystemDiskPercent))

		jobEphemeralDiskInodePercentMetric = prometheus.NewGaugeVec(
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

		jobEphemeralDiskInodePercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobEphemeralDiskInodePercent))

		jobEphemeralDiskPercentMetric = prometheus.NewGaugeVec(
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

		jobEphemeralDiskPercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobEphemeralDiskPercent))

		jobPersistentDiskInodePercentMetric = prometheus.NewGaugeVec(
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

		jobPersistentDiskInodePercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobPersistentDiskInodePercent))

		jobPersistentDiskPercentMetric = prometheus.NewGaugeVec(
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

		jobPersistentDiskPercentMetric.WithLabelValues(
			deploymentName,
			jobName,
			jobID,
			jobIndex,
		).Set(float64(jobPersistentDiskPercent))

		totalHMTSDBReceivedMessagesMetric = prometheus.NewCounter(
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

		totalHMTSDBInvalidMessagesMetric = prometheus.NewCounter(
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

		totalHMTSDBDiscardedMessagesMetric = prometheus.NewCounter(
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

		lastHMTSDBReceivedMessageTimestampMetric = prometheus.NewGauge(
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

		lastHMTSDBScrapeTimestampMetric = prometheus.NewGauge(
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

		lastHMTSDBScrapeDurationSecondsMetric = prometheus.NewGauge(
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

		hmTSDBCollector = NewHMTSDBCollector(namespace, environment, tsdbListener)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go hmTSDBCollector.Describe(descriptions)
		})

		It("returns a job_healthy metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobHealthyMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_load_avg01 metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobLoadAvg01Metric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_cpu_sys metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobCPUSysMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_cpu_user metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobCPUUserMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_cpu_wait metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobCPUWaitMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_mem_kb metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobMemKBMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_mem_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobMemPercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_swap_kb metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobSwapKBMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_swap_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobSwapPercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_system_disk_inode_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobSystemDiskInodePercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_system_disk_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobSystemDiskPercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_ephemeral_disk_inode_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobEphemeralDiskInodePercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_ephemeral_disk_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobEphemeralDiskPercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_persistent_disk_inode_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobPersistentDiskInodePercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a job_persistent_disk_percent metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(jobPersistentDiskPercentMetric.WithLabelValues(
				deploymentName,
				jobName,
				jobID,
				jobIndex,
			).Desc())))
		})

		It("returns a hm_tsdb_received_messages_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHMTSDBReceivedMessagesMetric.Desc())))
		})

		It("returns a hm_tsdb_invalid_messages_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHMTSDBInvalidMessagesMetric.Desc())))
		})

		It("returns a hm_tsdb_discarded_messages_total metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(totalHMTSDBDiscardedMessagesMetric.Desc())))
		})

		It("returns a hm_tsdb_last_received_message_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastHMTSDBReceivedMessageTimestampMetric.Desc())))
		})

		It("returns a last_hm_tsdb_scrape_timestamp metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastHMTSDBScrapeTimestampMetric.Desc())))
		})

		It("returns a last_hm_tsdb_scrape_duration_seconds metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(lastHMTSDBScrapeDurationSecondsMetric.Desc())))
		})
	})

	Describe("Collect", func() {
		var (
			conn    net.Conn
			metrics chan prometheus.Metric

			tsdbTags    string
			tsdbMessage string
		)

		BeforeEach(func() {
			metrics = make(chan prometheus.Metric)

			tsdbTags = fmt.Sprintf("deployment=%s job=%s index=%s id=%s", deploymentName, jobName, jobIndex, jobID)
		})

		JustBeforeEach(func() {
			conn, err = net.Dial("tcp", tsdbListener.Addr().String())
			Expect(err).ToNot(HaveOccurred())
			_, err = conn.Write([]byte(tsdbMessage))
			Expect(err).ToNot(HaveOccurred())
			conn.Close()

			// Leave some time to the tsdb parser to process the message
			time.Sleep(100 * time.Millisecond)
			go hmTSDBCollector.Collect(metrics)
		})

		Context("when a system.healthy message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.healthy %d 1 %s", time.Now().Unix(), tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_process_healthy metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobHealthyMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.load.1m message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.load.1m %d %f %s", time.Now().Unix(), jobLoadAvg01, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_load_avg01 metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobLoadAvg01Metric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.cpu.sys message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.cpu.sys %d %f %s", time.Now().Unix(), jobCPUSys, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_cpu_sys metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobCPUSysMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.cpu.user message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.cpu.user %d %f %s", time.Now().Unix(), jobCPUUser, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_cpu_user metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobCPUUserMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.cpu.wait message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.cpu.wait %d %f %s", time.Now().Unix(), jobCPUWait, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_cpu_wait metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobCPUWaitMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.mem.kb message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.mem.kb %d %d %s", time.Now().Unix(), jobMemKB, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_mem_kb metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobMemKBMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.mem.percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.mem.percent %d %d %s", time.Now().Unix(), jobMemPercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_mem_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobMemPercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.swap.kb message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.swap.kb %d %d %s", time.Now().Unix(), jobSwapKB, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_swap_kb metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobSwapKBMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.swap.percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.swap.percent %d %d %s", time.Now().Unix(), jobSwapPercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_swap_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobSwapPercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.disk.system.inode_percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.disk.system.inode_percent %d %d %s", time.Now().Unix(), jobSystemDiskInodePercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_system_disk_inode_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobSystemDiskInodePercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.disk.system.percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.disk.system.percent %d %d %s", time.Now().Unix(), jobSystemDiskPercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_system_disk_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobSystemDiskPercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.disk.ephemeral.inode_percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.disk.ephemeral.inode_percent %d %d %s", time.Now().Unix(), jobEphemeralDiskInodePercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_ephemeral_disk_inode_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobEphemeralDiskInodePercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.disk.ephemeral.percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.disk.ephemeral.percent %d %d %s", time.Now().Unix(), jobEphemeralDiskPercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_ephemeral_disk_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobEphemeralDiskPercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.disk.persistent.inode_percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.disk.persistent.inode_percent %d %d %s", time.Now().Unix(), jobPersistentDiskInodePercent, tsdbTags)
				totalHMTSDBReceivedMessagesMetric.Inc()
			})

			It("returns a job_persistent_disk_inode_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobPersistentDiskInodePercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when a system.disk.persistent.percent message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put system.disk.persistent.percent %d %d %s", time.Now().Unix(), jobPersistentDiskPercent, tsdbTags)
			})

			It("returns a job_persistent_disk_percent metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(jobPersistentDiskPercentMetric.WithLabelValues(
					deploymentName,
					jobName,
					jobID,
					jobIndex,
				))))
			})
		})

		Context("when an invalid tsdb message is received", func() {
			Context("when does not have the right number of tokens", func() {
				BeforeEach(func() {
					tsdbMessage = fmt.Sprintf("put invalid.tsdb.message %d", time.Now().Unix())
					totalHMTSDBInvalidMessagesMetric.Inc()
				})

				It("returns a hm_tsdb_invalid_messages_total metric metric", func() {
					Eventually(metrics).Should(Receive(PrometheusMetric(totalHMTSDBInvalidMessagesMetric)))
				})
			})

			Context("when the value cannot be converted to a float", func() {
				BeforeEach(func() {
					tsdbMessage = fmt.Sprintf("put invalid.tsdb.message %d a %s", time.Now().Unix(), tsdbTags)
					totalHMTSDBInvalidMessagesMetric.Inc()
				})

				It("returns a hm_tsdb_invalid_messages_total metric metric", func() {
					Eventually(metrics).Should(Receive(PrometheusMetric(totalHMTSDBInvalidMessagesMetric)))
				})
			})
		})

		Context("when a non supported tsdb message is received", func() {
			BeforeEach(func() {
				tsdbMessage = fmt.Sprintf("put invalid.tsdb.message %d 1 %s", time.Now().Unix(), tsdbTags)
				totalHMTSDBDiscardedMessagesMetric.Inc()
			})

			It("returns a hm_tsdb_discarded_messages_total metric", func() {
				Eventually(metrics).Should(Receive(PrometheusMetric(totalHMTSDBDiscardedMessagesMetric)))
			})
		})
	})
})
