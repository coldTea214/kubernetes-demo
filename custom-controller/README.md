官方文档：[operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

custom controller 一般不操作内部类型，而跟 CRD 交互，通常作为 operator 的实现机制之一

以前录过相关的课程，当时是用的 kubebuilder 工具，就先放这，暂不写其它 demo 了：https://edu.aliyun.com/lesson_1651_13095?spm=5176.254948.1387840.29.2c12cad2GJGB3D#_13095