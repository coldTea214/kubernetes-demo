官方文档：[scheduler extender](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/scheduler_extender.md)

演示：

* 部署 deployment

	```
	$ kubectl apply -f scheduler-extender.yaml
	```
* 调整 kube-scheduler，配置上此 extender。配置的方法很多，这里给出其中一种：kube-scheduler 增加启动参数 ```--policy-config-file=policy.cfg```
* 确认下 scheduler 策略已生效

	```
	$ kubectl -n kube-system logs kube-scheduler-n227-020-128 | grep extender
	I0911 08:27:44.766434       1 factory.go:346] Creating extender with config ...
	```
* 正常创建 pod，就可以验证了	

注意，因为只是 demo 项目，并没有特意做性能调优，生产环境使用的话，可以：

* policy.cfg 配置下 ManagedResources，减少不必要的 http 调用
* policy.cfg 配置下 NodeCacheCapable，extender 自己来维护 node cache，减少 http 调用传递的数据