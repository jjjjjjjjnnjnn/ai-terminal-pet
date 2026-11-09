你现在是在**修订**一份已经写好的中文稿，不是重写。先完整读一遍 drafts/README.zh.md。这份稿子整体可用，你只做下面八处定点修改。没被点名的段落一个字都不要动；不要换结构，不要合并小节，不要重新润色已经正确的句子。改完的稿子必须仍看得出是同一篇。

【已由仓库作者核实的事实，以此为准】
- 画布是纯 GDI，不是 GDI+。internal/ui/canvas_windows.go:13 只引入 syscall 和 github.com/lxn/win；:61 CreateCompatibleDC，:75 CreateDIBSection（32bpp、BI_RGB、biHeight 取负数即自顶向下 DIB），:80 SelectObject，最后 UpdateLayeredWindow 交给分层窗口。全仓没有任何 gdiplus。AGENTS.md:25 那句“GDI+ canvas rasterization”是过时错误描述，不要引用它。
- 命名管道的 DACL 字面量是 "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)"（internal/pipeserver/server_windows.go:68）。含义：只给 Owner Rights（OW，即运行程序的当前 Windows 用户）和 LocalSystem（SY）读写权限，本机其他账户既连不上这条管道，也抢不走这个管道名。
- internal/session/session.go:17 MaxPayloadBytes = 1 << 20（1 MiB）；:16 SchemaVersion 为 "ai-terminal-pet.v1"；:21-29 定义七状态 idle/thinking/working/tool_use/waiting_permission/error/closed。
- internal/pipeserver/server_windows.go:23 是 32 路并发连接；cmd/pet-statusline/main.go:375 的回执就是精确字符串 "OK"。
- store.DefaultAbsenceGrace = 2 分钟（internal/store/store.go:22）；无会话自毁在 internal/app/app_windows.go:830：time.Since(a.lastAvailable) >= 10*time.Minute。

【八处必改】

1）第 3 行定位句现在是“常驻屏幕边缘的 Windows 状态浮窗，一眼掌握多路 AI 终端会话”，一句里塞了两件事。拆成两句、每句不超过 20 个字：第一句说它替你消掉了什么动作（不必再逐个终端窗口来回扫），第二句说它的代价（一个可执行文件、不要账号、不联网）。规格数字（体积、像素）不要留在这一行，挪到下面的徽章或正文。

2）功能表里“脚本控制与单实例”那一行把四件互不相干的事塞进了一格。删掉它，补三行，每行只讲一件事：
  - “单实例运行”：具名互斥体保证第二次启动只是唤醒已经运行的浮窗，不会再起一个进程。
  - “连接容量”：管道服务最多接受 32 路并发连接，读超时 5 秒，IPC 不占用 UI 线程。
  - “空闲生命周期”：宿主连续 2 分钟不可观测才移除会话；一个会话都不剩时，进程在 10 分钟后自行退出。
  保持表格原有列数和行文口吻。

3）架构图里那一行“Win32 分层浮窗 (置顶渲染)”要写明是 GDI 画布（改成类似“Win32 分层浮窗 (GDI 置顶画布)”）。然后在图下的原理列表里补一条说明它怎么画：用 CreateDIBSection 建立自顶向下的 32bpp DIB，通过内存 DC 绘制，再交给 UpdateLayeredWindow 贴到分层窗口。一句话讲清顺序即可，不要写成吹点。

4）安装段里删掉第 1 步“静态检查与单元测试 go vet / go test”。装产品的人不是来当贡献者的。安装只剩三步：编译两个二进制 → 跑 .\install.ps1 → 用 ai-terminal-pet --json 验证。然后在文末、许可证之后新增一节“## 从源码开发”，把验证命令放进去：go vet ./...、go test -count=1 ./...、node --check reporters/*.js、node integrations/deepseek-harness/verify.mjs，以及 PET_PREVIEW_DIR 预览命令。这一节是给改代码的人看的，前面加一句交代即可。

5）接入表里 DeepSeek Harness 那行的“得到什么”现在写的是“官方 Cordis 事件总线接入，自动映射会话状态与错误”，太泛且含猜测。改成真实事件映射：turn/start→working、tool/call→tool_use、approval/asked→waiting_permission、tool/result 出错→error、turn/end→idle、进程退出→closed。并在表下补一句：Harness 那一侧是 host-only 的 Cordis bundle，自身不含管道名、状态集与快照逻辑，它从 %LOCALAPPDATA%\AI-Terminal-Pet\protocol.js 加载协议，全部委托给官方 protocol.js。

6）安全边界里“命名管道按用户隔离”只说了“收紧 ACL 访问控制权限”，太空。改成具体：管道 DACL 为 "D:P(A;;GRGW;;;OW)(A;;GRGW;;;SY)"，只授予管道属主（运行浮窗的那个 Windows 用户）与本机 LocalSystem 读写权限，本机其他账户既不能连接也不能抢占这个管道名。

7）路线图里“二进制体积压缩”写的是“两个可执行文件的总体积控制在 5 MiB 以内”。5 MiB 合计这个说法路线图里没有，删掉。docs/ROADMAP.md 的 L2 原文是：现状 3.7 MiB，目标主程序 <2.5 MiB。只保留一个目标：主程序小于 2.5 MiB。手段可简述（剥离无用反射元数据、去掉未用包、把 pet-statusline 里基于 regexp 的脱敏规则改成手写扫描器）。

8）“非目标（明确不做）”这一节现在就两条，结构是对的，保留这个形态，只做两件事：确认它确实是“## 非目标（明确不做）”这一节下的两条子项（不要再多出并列的“非目标”条目），并且两条的措辞要落到底：① 不做账号体系、不做云同步、不做遥测，数据只落在 %LOCALAPPDATA%；② 不是终端复用器，也不替代你的终端，它只回答“哪个会话在等你”。

【本次修订的硬规矩】
- 除这八处，别的地方不要动。小节顺序、图片文件名、表格列名一律不变。
- 禁用词继续禁：革命性、颠覆、极致、丝滑、强大、优雅、赋能、神器、黑科技、遥遥领先、一键搞定、体验拉满、效率翻倍；英文那一套（revolutionary/blazingly/seamless/effortless 等）同样不许出现。
- 全文不许有感叹号，标题里不许有 emoji。
- 中文稿是重写不是直译：句子要像中文技术博客作者写的，不要出现“通过……来……”式的翻译腔堆叠。你只改八处，改动处的措辞请用自然中文，别退回机翻味。
- 每个数字都必须是上面核实过的、或原稿里已经有的。不许自己造数。
- 篇幅控制在 200–260 行。你要是把没被点名的小节改短或改长了，就是改错了。

【输出格式】把修订后的完整 Markdown 直接输出到 stdout，不要任何前言、不要“以下是修订版本”这类话、不要用代码围栏包住全文。第一行就是 "# ai-terminal-pet"。