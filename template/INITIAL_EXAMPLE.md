## FEATURE:

- 具有另一个Pydantic AI代理作为工具的Pydantic AI代理。
- 主代理的研究代理，然后是子代理的电子邮件草稿代理。
- 与代理交互的CLI。
- 电子邮件草稿代理使用Gmail，研究代理使用Brave API。

## EXAMPLES:

在 `examples/` 文件夹中，有一个README供你阅读，以了解示例的全部内容，以及当你为上述功能创建文档时如何构建自己的README。

- `examples/cli.py` - 使用这个作为创建CLI的模板
- `examples/agent/` - 阅读这里的所有文件，了解创建支持不同提供者和LLM、处理代理依赖关系以及向代理添加工具的Pydantic AI代理的最佳实践。

不要直接复制这些示例中的任何一个，它是为完全不同的项目准备的。但请将其用作灵感和最佳实践的参考。

## DOCUMENTATION:

Pydantic AI文档：https://ai.pydantic.dev/

## OTHER CONSIDERATIONS:

- 包含一个.env.example，以及包含设置说明的README，包括如何配置Gmail和Brave。
- 在README中包含项目结构。
- 虚拟环境已经设置好，包含了必要的依赖项。
- 使用python_dotenv和load_env()处理环境变量
