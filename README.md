# xmind
基于go语言的xmind接口

使用方法参考: [example](example)

本库主要加载xmind文件为json结构,保存文件时也用的json结构而不是xml结构

本库只做了最基本的主题添加功能,类似`标签/备注/图片`等其他功能不考虑,有想法的自行实现

本库做了通用加载和通用保存方法,可以更灵活的与其他思维导图进行转换

参考: [custom_test](example/custom_test.go)
