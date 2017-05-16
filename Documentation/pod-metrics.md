# Pod Metrics

| Metric name| Metric type | Labels/tags | Examples   |
| ---------- | ----------- | ----------- |----------- |
| kube_pod_info | Gauge | `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `host_ip`=&lt;host-ip&gt; <br> `pod_ip`=&lt;pod-ip&gt; <br> `start_time`=&lt;date-time since kubelet acknowledged pod&gt; <br> `node`=&lt;node-name&gt;<br> `created_by`=&lt;created_by&gt;  |kube_pod_info{host_ip="10.142.21.158",namespace="kube-system",pod="nginx-ingress-controller-49379976-bsw1p",pod_ip="10.142.21.158"} 1|
| kube_pod_labels | Gauge | `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `label_POD_LABEL`=&lt;POD_LABEL&gt;  |
| kube_pod_status_phase | Gauge | `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `phase`=&lt;Pending\|Running\|Succeeded\|Failed\|Unknown&gt; |
| kube_pod_status_ready | Gauge |  `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `condition`=&lt;true\|false\|unknown&gt; |
| kube_pod_status_scheduled | Gauge |  `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `condition`=&lt;true\|false\|unknown&gt; |
| kube_pod_container_info | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `image`=&lt;image-name&gt; <br> `image_id`=&lt;image-id&gt; <br> `container_id`=&lt;containerid&gt; |
| kube_pod_container_status_waiting | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; |
| kube_pod_container_status_running | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; |
| kube_pod_container_status_terminated | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; |
| kube_pod_container_status_ready | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; |
| kube_pod_container_status_restarts | Counter | `container`=&lt;container-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `pod`=&lt;pod-name&gt; |
| kube_pod_container_resource_requests_cpu_cores | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |
| kube_pod_container_resource_requests_memory_bytes | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |
| kube_pod_container_resource_limits_cpu_cores | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |
| kube_pod_container_resource_limits_memory_bytes | Gauge | `container`=&lt;container-name&gt; <br> `pod`=&lt;pod-name&gt; <br> `namespace`=&lt;pod-namespace&gt; <br> `node`=&lt; node-name&gt; |
