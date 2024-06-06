# feishu-openai

<p align='center'>
   飞书 × openai 
<br>
<br>
    🚀 Feishu OpenAI 🚀
</p>

## 👻 机器人功能
1. 文字聊天
2. 文字+图片聊天(vision)
3. 文生图 /pic + prompt
4. SpeechToText
5. TextToSpech

## 🌟 项目特点

- 🍏 对话基于 OpenAI(https://platform.openai.com/account/api-keys) 接口
- 🍎 通过 lark，将 ChatGPT 接入[飞书](https://open.feishu.cn/app)
- 基于飞书长连接事件回调，不需要公网IP
- 基于飞书消息更新，流式回复
- 支持Azure

## 项目部署

### OpenAI部署

```bash
docker run -d --restart=always --name feishu-openai \
--env FEISHUAPP_ID=xxx \
--env FEISHUAPP_SECRET=xxx \
--env FEISHU_ENCRYPT_KEY=xxx \
--env FEISHU_VERIFICATION_TOKEN=xxx \
--env OPENAI_MODEL=gpt-4o \
--env OPENAI_KEY=sk-xxx1 \
blacklee123/feishu-openai:latest
```

### Azure部署

```bash
docker run -d --restart=always --name feishu-openai \
--env FEISHUAPP_ID=xxx \
--env FEISHUAPP_SECRET=xxx \
--env FEISHU_ENCRYPT_KEY=xxx \
--env FEISHU_VERIFICATION_TOKEN=xxx \
--env OPENAI_MODEL=gpt-4o \
--env AZURE_ON=true \
--env AZURE_ENDPOINT=your_end_point \
--env AZURE_DEPLOYMENT_NAME=gpt-4o \
--env AZURE_OPENAI_TOKEN=your_token
blacklee123/feishu-openai:latest
```

## 详细配置步骤



- 获取 [OpenAI](https://platform.openai.com/account/api-keys) 的 KEY
- 创建 [飞书](https://open.feishu.cn/) 机器人
    1. 前往[开发者平台](https://open.feishu.cn/app?lang=zh-CN)创建应用,并获取到 APPID 和 Secret
    2. 前往`应用功能-机器人`, 创建机器人
    3. 在事件订阅板块，搜索 `接收消息`, 把他们后面所有的权限全部勾选。
       进入权限管理界面，搜索`图片`, 勾选`获取与上传图片或文件资源`。
       最终会添加下列权限
        - contact:contact.base:readonly(获取通讯录基本信息)
        - contact:user.base:readonly(获取用户基本信息)
        - im:resource(获取与上传图片或文件资源)
        - im:message
        - im:message.group_at_msg:readonly(接收群聊中@机器人消息事件)
        - im:message.p2p_msg(获取用户发给机器人的单聊消息)
        - im:message.p2p_msg:readonly(读取用户发给机器人的单聊消息)
        - im:message:send_as_bot(获取用户在群组中@机器人的消息)
    4. 发布版本，等待企业管理员审核通过

## 加入答疑群

[单击加入答疑群](https://applink.feishu.cn/client/chat/chatter/add_by_link?link_token=c88k15ff-64d9-4d7a-8b3c-f19750764c7c)

<img src='./docs/talk.png' alt='' width='200'/>
