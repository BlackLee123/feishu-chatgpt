<p align='center'>
    <img src='./docs/img3.png' alt='' width='800'/>
</p>

<details align='center'>
    <summary> 📷 点击展开完整功能截图</summary>
    <br>
    <p align='center'>
    <img src='https://user-images.githubusercontent.com/50035229/224493411-085ba405-81cd-4972-b87b-74a2e811f23d.png' alt='语音对话' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/223590381-ed38db74-39f3-4e77-bd3a-aaa54d679286.png' alt='角色扮演' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/228092540-bb9c74de-fccb-4d07-9797-f3fe1fb3705a.png' alt='角色列表' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/223590817-37a56eac-ab6e-4293-862a-de0988ac50b7.png' alt='文字成图' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/224531308-378a5bc3-2964-4e32-905f-8237dafc3f91.png' alt='图片变体' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/227900488-10a53176-f4e3-4b9e-bf9e-b60d958aefc8.png' alt='余额查询' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/224531775-3f0e1e1b-a373-4774-a8f0-e120ccba6670.png' alt='帮助菜单' width='800'/>
    </p>

</details>

<br>

<p align='center'>
   飞书 × openai × 科大讯飞
<br>
<br>
    🚀 Feishu OpenAI 🚀
</p>

## 👻 机器人功能

🗣 语音交流：私人直接与机器人畅所欲言

💬 多话题对话：支持私人和群聊多话题讨论，高效连贯

🖼 文本成图：支持文本成图和以图搜图

🛖 场景预设：内置丰富场景列表，一键切换AI角色

🎭 角色扮演：支持场景模式，增添讨论乐趣和创意

🔄 上下文保留：回复对话框即可继续同一话题讨论

⏰ 自动结束：超时自动结束对话，支持清除讨论历史

📝 富文本卡片：支持富文本卡片回复，信息更丰富多彩

👍 交互式反馈：即时获取机器人处理结果

🎰 余额查询：即时获取token消耗情况

🔙 历史回档：轻松回档历史对话，继续话题讨论 🚧

🔒 管理员模式：内置管理员模式，使用更安全可靠 🚧

🌐 多token负载均衡：优化生产级别的高频调用场景

↩️ 支持反向代理：为不同地区的用户提供更快、更稳定的访问体验

📚 与飞书文档互动：成为企业员工的超级助手 🚧

🎥 话题内容秒转PPT：让你的汇报从此变得更加简单 🚧

📊 表格分析：轻松导入飞书表格，提升数据分析效率 🚧

🍊 私有数据训练：利用公司产品信息对GPT二次训练，更好地满足客户个性化需求 🚧

## 🌟 项目特点

- 🍏 对话基于 OpenAI(https://platform.openai.com/account/api-keys) 接口
- 🍎 通过 lark，将 ChatGPT 接入[飞书](https://open.feishu.cn/app)和[飞书国际版](https://www.larksuite.com/)
- 🥒[Docker](https://www.docker.com/)部署
- 🍋 基于[goCache](https://github.com/patrickmn/go-cache)内存键值对缓存

## 项目部署

<details>
    <summary>docker部署</summary>
<br>

```bash
docker build -t feishu-chatgpt:latest .
docker run -d --name feishu-chatgpt -p 9000:9000 \
--env APP_ID=xxx \
--env APP_SECRET=xxx \
--env APP_ENCRYPT_KEY=xxx \
--env APP_VERIFICATION_TOKEN=xxx \
--env BOT_NAME=chatGpt \
--env OPENAI_MODEL="gpt-4" \
--env OPENAI_KEY="sk-xxx1,sk-xxx2,sk-xxx3" \
--env API_URL="https://api.openai.com" \
--env HTTP_PROXY="" \
--env OPENAI_KEY="sk-xxx1,sk-xxx2,sk-xxx3" \
--env API_URL="https://api.openai.com" \
--env HTTP_PROXY="" \
--env XFAPP_ID="科大讯飞开发平台APPID;不需要语音回复可不填写" \
--env XFAPP_SECRET="科大讯飞开发平台APISecret;不需要语音回复可不填写" \
--env XFAPP_KEY="科大讯飞开发平台APIKey;不需要语音回复可不填写" \
--env OpenaiMaxTokens="2000" \
--env STREAM_MODE="true" \
blacklee123/feishu-chatgpt:latest
```

注意:

- `BOT_NAME` 为飞书机器人名称，例如 `chatGpt`
- `OPENAI_KEY` 为openai key，多个key用逗号分隔，例如 `sk-xxx1,sk-xxx2,sk-xxx3`
- `HTTP_PROXY` 为宿主机的proxy地址，例如 `http://host.docker.internal:7890`,没有代理的话，可以不用设置
- `API_URL` 为openai api 接口地址，例如 `https://api.openai.com`, 没有反向代理的话，可以不用设置

---

事件回调地址: http://IP:9000/webhook/event
卡片回调地址: http://IP:9000/webhook/card

把它填入飞书后台
<br>

</details>

## 详细配置步骤

<details align='left'>
    <summary> 📸 点击展开飞书机器人配置的分步截图指导</summary>
    <br>
    <p align='center'>
    <img src='https://user-images.githubusercontent.com/50035229/223943381-39e0466f-2a5e-472a-9863-94eafb5f17b0.png' alt='' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/223943448-228de5cb-0929-4d80-8087-8d8624dd6ddf.png' alt='' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/223943485-ef331784-7940-4657-b128-70c98391e72f.png' alt='' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/223943527-60e6653a-eb6e-4062-a076-b6c9da934352.png' alt='' width='800'/>
    <img src='https://user-images.githubusercontent.com/50035229/223943972-f49adf9f-af5f-463a-8c7a-c1f0cac0e8c3.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944060-7ef630a4-4248-4509-852b-cad8bfffeefc.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944230-aff586be-31cc-40de-9b1a-7d4e259d54dd.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944350-917d115c-6c82-4d8b-9ec8-b5c82331a2dc.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944381-97396156-f5e2-467f-aaf6-b1f6e1c446b2.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223945122-f7ab3d9a-6742-43d2-970e-ddb0f284c7fa.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944507-8d1a08d7-8b5b-4f32-a90d-fd338164ec82.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944515-fb505e84-c840-484a-8df5-612f60bf27ea.png' alt='' width='800'/>
      <img src='https://user-images.githubusercontent.com/50035229/223944590-ad61320f-c14a-4542-80ad-dee2e6469b67.png' alt='' width='800'/>
    </p>
</details>

- 获取 [OpenAI](https://platform.openai.com/account/api-keys) 的 KEY
- 创建 [飞书](https://open.feishu.cn/) 机器人
    1. 前往[开发者平台](https://open.feishu.cn/app?lang=zh-CN)创建应用,并获取到 APPID 和 Secret
    2. 前往`应用功能-机器人`, 创建机器人
    3. 从 cpolar、serverless 或 Railway 获得公网地址，在飞书机器人后台的 `事件订阅` 板块填写。例如，
        - `http://xxxx.r6.cpolar.top`为 cpolar 暴露的公网地址
        - `/webhook/event`为统一的应用路由
        - 最终的回调地址为 `http://xxxx.r6.cpolar.top/webhook/event`
    4. 在飞书机器人后台的 `机器人` 板块，填写消息卡片请求网址。例如，
        - `http://xxxx.r6.cpolar.top`为 cpolar 暴露的公网地址
        - `/webhook/card`为统一的应用路由
        - 最终的消息卡片请求网址为 `http://xxxx.r6.cpolar.top/webhook/card`
    5. 在事件订阅板块，搜索三个词`机器人进群`、 `接收消息`、 `消息已读`, 把他们后面所有的权限全部勾选。
       进入权限管理界面，搜索`图片`, 勾选`获取与上传图片或文件资源`。
       最终会添加下列回调事件
        - im:resource(获取与上传图片或文件资源)
        - im:message
        - im:message.group_at_msg(获取群组中所有消息)
        - im:message.group_at_msg:readonly(接收群聊中@机器人消息事件)
        - im:message.p2p_msg(获取用户发给机器人的单聊消息)
        - im:message.p2p_msg:readonly(读取用户发给机器人的单聊消息)
        - im:message:send_as_bot(获取用户在群组中@机器人的消息)
        - im:chat:readonly(获取群组信息)
        - im:chat(获取与更新群组信息)
    6. 发布版本，等待企业管理员审核通过

## 加入答疑群

[单击加入答疑群](https://applink.feishu.cn/client/chat/chatter/add_by_link?link_token=c88k15ff-64d9-4d7a-8b3c-f19750764c7c)

<img src='./docs/talk.png' alt='' width='200'/>
