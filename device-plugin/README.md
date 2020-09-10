device plugin demo 的功能：将 /mnt/file-devices 目录下的文件作为 "设备" 使用

演示：

* 在宿主机端创建 /mnt/file-devices 目录，并随便创建文件

	```
	$ ls /mnt/file-devices/
	device1  device2  device3
	```
* 创建 daemonset

	```
	$ kubectl apply -f ds.yaml
	```
* 确认 "file 设备" 已正常上报

	```
	$ kubectl get node n227-020-128 -o yaml | grep coldtea
	    coldtea214/file: "3"
	    coldtea214/file: "3"
	```
* 创建使用 "file 设备" 的 pod

	```
	$ kubectl apply -f pod.yaml
	```
* 验证 pod 申请到了所需的 "设备"

	```
	$ kubectl exec -it demo-pod sh
	/ # ls /mnt/file-devices/
	device1  device3
	```