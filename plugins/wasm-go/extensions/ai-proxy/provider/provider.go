package provider

import (
	"bytes"
	"errors"
	"math/rand"
	"net/http"
	"path"
	"regexp"
	"strconv"

	"strings"

	"github.com/alibaba/higress/plugins/wasm-go/extensions/ai-proxy/util"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/higress-group/wasm-go/pkg/log"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type (
	ApiName          string
	Pointcut         string
	basePathHandling string
)

const (

	// ApiName 格式 {vendor}/{version}/{apitype}
	// 表示遵循 厂商/版本/接口类型 的格式
	// 目前openai是事实意义上的标准，但是也有其他厂商存在其他任务的一些可能的标准，比如cohere的rerank
	ApiNameCompletion                           ApiName = "openai/v1/completions"
	ApiNameChatCompletion                       ApiName = "openai/v1/chatcompletions"
	ApiNameEmbeddings                           ApiName = "openai/v1/embeddings"
	ApiNameImageGeneration                      ApiName = "openai/v1/imagegeneration"
	ApiNameImageEdit                            ApiName = "openai/v1/imageedit"
	ApiNameImageVariation                       ApiName = "openai/v1/imagevariation"
	ApiNameAudioSpeech                          ApiName = "openai/v1/audiospeech"
	ApiNameFiles                                ApiName = "openai/v1/files"
	ApiNameRetrieveFile                         ApiName = "openai/v1/retrievefile"
	ApiNameRetrieveFileContent                  ApiName = "openai/v1/retrievefilecontent"
	ApiNameBatches                              ApiName = "openai/v1/batches"
	ApiNameRetrieveBatch                        ApiName = "openai/v1/retrievebatch"
	ApiNameCancelBatch                          ApiName = "openai/v1/cancelbatch"
	ApiNameModels                               ApiName = "openai/v1/models"
	ApiNameResponses                            ApiName = "openai/v1/responses"
	ApiNameFineTuningJobs                       ApiName = "openai/v1/fine-tuningjobs"
	ApiNameRetrieveFineTuningJob                ApiName = "openai/v1/retrievefine-tuningjob"
	ApiNameFineTuningJobEvents                  ApiName = "openai/v1/fine-tuningjobsevents"
	ApiNameFineTuningJobCheckpoints             ApiName = "openai/v1/fine-tuningjobcheckpoints"
	ApiNameCancelFineTuningJob                  ApiName = "openai/v1/cancelfine-tuningjob"
	ApiNameResumeFineTuningJob                  ApiName = "openai/v1/resumefine-tuningjob"
	ApiNamePauseFineTuningJob                   ApiName = "openai/v1/pausefine-tuningjob"
	ApiNameFineTuningCheckpointPermissions      ApiName = "openai/v1/fine-tuningjobcheckpointpermissions"
	ApiNameDeleteFineTuningCheckpointPermission ApiName = "openai/v1/deletefine-tuningjobcheckpointpermission"

	// TODO: 以下是一些非标准的API名称，需要进一步确认是否支持
	ApiNameCohereV1Rerank              ApiName = "cohere/v1/rerank"
	ApiNameQwenAsyncAIGC               ApiName = "qwen/v1/services/aigc"
	ApiNameQwenAsyncTask               ApiName = "qwen/v1/tasks"
	ApiNameQwenV1Rerank                ApiName = "qwen/v1/rerank"
	ApiNameGeminiGenerateContent       ApiName = "gemini/v1beta/generatecontent"
	ApiNameGeminiStreamGenerateContent ApiName = "gemini/v1beta/streamgeneratecontent"
	ApiNameAnthropicMessages           ApiName = "anthropic/v1/messages"
	ApiNameAnthropicComplete           ApiName = "anthropic/v1/complete"

	// OpenAI
	PathOpenAICompletions                          = "/v1/completions"
	PathOpenAIChatCompletions                      = "/v1/chat/completions"
	PathOpenAIEmbeddings                           = "/v1/embeddings"
	PathOpenAIFiles                                = "/v1/files"
	PathOpenAIRetrieveFile                         = "/v1/files/{file_id}"
	PathOpenAIRetrieveFileContent                  = "/v1/files/{file_id}/content"
	PathOpenAIBatches                              = "/v1/batches"
	PathOpenAIRetrieveBatch                        = "/v1/batches/{batch_id}"
	PathOpenAICancelBatch                          = "/v1/batches/{batch_id}/cancel"
	PathOpenAIModels                               = "/v1/models"
	PathOpenAIImageGeneration                      = "/v1/images/generations"
	PathOpenAIImageEdit                            = "/v1/images/edits"
	PathOpenAIImageVariation                       = "/v1/images/variations"
	PathOpenAIAudioSpeech                          = "/v1/audio/speech"
	PathOpenAIResponses                            = "/v1/responses"
	PathOpenAIFineTuningJobs                       = "/v1/fine_tuning/jobs"
	PathOpenAIRetrieveFineTuningJob                = "/v1/fine_tuning/jobs/{fine_tuning_job_id}"
	PathOpenAIFineTuningJobEvents                  = "/v1/fine_tuning/jobs/{fine_tuning_job_id}/events"
	PathOpenAIFineTuningJobCheckpoints             = "/v1/fine_tuning/jobs/{fine_tuning_job_id}/checkpoints"
	PathOpenAICancelFineTuningJob                  = "/v1/fine_tuning/jobs/{fine_tuning_job_id}/cancel"
	PathOpenAIResumeFineTuningJob                  = "/v1/fine_tuning/jobs/{fine_tuning_job_id}/resume"
	PathOpenAIPauseFineTuningJob                   = "/v1/fine_tuning/jobs/{fine_tuning_job_id}/pause"
	PathOpenAIFineTuningCheckpointPermissions      = "/v1/fine_tuning/checkpoints/{fine_tuned_model_checkpoint}/permissions"
	PathOpenAIFineDeleteTuningCheckpointPermission = "/v1/fine_tuning/checkpoints/{fine_tuned_model_checkpoint}/permissions/{permission_id}"

	// Anthropic
	PathAnthropicMessages = "/v1/messages"
	PathAnthropicComplete = "/v1/complete"

	// Cohere
	PathCohereV1Rerank = "/v1/rerank"

	providerTypeMoonshot   = "moonshot"
	providerTypeAzure      = "azure"
	providerTypeAi360      = "ai360"
	providerTypeGithub     = "github"
	providerTypeQwen       = "qwen"
	providerTypeOpenAI     = "openai"
	providerTypeGroq       = "groq"
	providerTypeBaichuan   = "baichuan"
	providerTypeYi         = "yi"
	providerTypeDeepSeek   = "deepseek"
	providerTypeZhipuAi    = "zhipuai"
	providerTypeOllama     = "ollama"
	providerTypeClaude     = "claude"
	providerTypeBaidu      = "baidu"
	providerTypeHunyuan    = "hunyuan"
	providerTypeStepfun    = "stepfun"
	providerTypeMinimax    = "minimax"
	providerTypeCloudflare = "cloudflare"
	providerTypeSpark      = "spark"
	providerTypeGemini     = "gemini"
	providerTypeDeepl      = "deepl"
	providerTypeMistral    = "mistral"
	providerTypeCohere     = "cohere"
	providerTypeDoubao     = "doubao"
	providerTypeCoze       = "coze"
	providerTypeTogetherAI = "together-ai"
	providerTypeDify       = "dify"
	providerTypeBedrock    = "bedrock"
	providerTypeVertex     = "vertex"

	protocolOpenAI   = "openai"
	protocolOriginal = "original"

	roleSystem    = "system"
	roleAssistant = "assistant"
	roleUser      = "user"

	finishReasonStop   = "stop"
	finishReasonLength = "length"

	ctxKeyIncrementalStreaming   = "incrementalStreaming"
	ctxKeyApiKey                 = "apiKey"
	CtxKeyApiName                = "apiName"
	ctxKeyIsStreaming            = "isStreaming"
	ctxKeyStreamingBody          = "streamingBody"
	ctxKeyOriginalRequestModel   = "originalRequestModel"
	ctxKeyFinalRequestModel      = "finalRequestModel"
	ctxKeyPushedMessage          = "pushedMessage"
	ctxKeyContentPushed          = "contentPushed"
	ctxKeyReasoningContentPushed = "reasoningContentPushed"

	objectChatCompletion      = "chat.completion"
	objectChatCompletionChunk = "chat.completion.chunk"

	reasoningBehaviorPassThrough = "passthrough"
	reasoningBehaviorIgnore      = "ignore"
	reasoningBehaviorConcat      = "concat"

	wildcard = "*"

	defaultTimeout = 2 * 60 * 1000 // ms

	basePathHandlingRemovePrefix basePathHandling = "removePrefix"
	basePathHandlingPrepend      basePathHandling = "prepend"
)

type providerInitializer interface {
	ValidateConfig(*ProviderConfig) error
	CreateProvider(ProviderConfig) (Provider, error)
}

var (
	errUnsupportedApiName = errors.New("unsupported API name")

	providerInitializers = map[string]providerInitializer{
		providerTypeMoonshot:   &moonshotProviderInitializer{},
		providerTypeAzure:      &azureProviderInitializer{},
		providerTypeAi360:      &ai360ProviderInitializer{},
		providerTypeGithub:     &githubProviderInitializer{},
		providerTypeQwen:       &qwenProviderInitializer{},
		providerTypeOpenAI:     &openaiProviderInitializer{},
		providerTypeGroq:       &groqProviderInitializer{},
		providerTypeBaichuan:   &baichuanProviderInitializer{},
		providerTypeYi:         &yiProviderInitializer{},
		providerTypeDeepSeek:   &deepseekProviderInitializer{},
		providerTypeZhipuAi:    &zhipuAiProviderInitializer{},
		providerTypeOllama:     &ollamaProviderInitializer{},
		providerTypeClaude:     &claudeProviderInitializer{},
		providerTypeBaidu:      &baiduProviderInitializer{},
		providerTypeHunyuan:    &hunyuanProviderInitializer{},
		providerTypeStepfun:    &stepfunProviderInitializer{},
		providerTypeMinimax:    &minimaxProviderInitializer{},
		providerTypeCloudflare: &cloudflareProviderInitializer{},
		providerTypeSpark:      &sparkProviderInitializer{},
		providerTypeGemini:     &geminiProviderInitializer{},
		providerTypeDeepl:      &deeplProviderInitializer{},
		providerTypeMistral:    &mistralProviderInitializer{},
		providerTypeCohere:     &cohereProviderInitializer{},
		providerTypeDoubao:     &doubaoProviderInitializer{},
		providerTypeCoze:       &cozeProviderInitializer{},
		providerTypeTogetherAI: &togetherAIProviderInitializer{},
		providerTypeDify:       &difyProviderInitializer{},
		providerTypeBedrock:    &bedrockProviderInitializer{},
		providerTypeVertex:     &vertexProviderInitializer{},
	}
)

type Provider interface {
	GetProviderType() string
}

type RequestHeadersHandler interface {
	OnRequestHeaders(ctx wrapper.HttpContext, apiName ApiName) error
}

type RequestBodyHandler interface {
	OnRequestBody(ctx wrapper.HttpContext, apiName ApiName, body []byte) (types.Action, error)
}

type StreamingResponseBodyHandler interface {
	OnStreamingResponseBody(ctx wrapper.HttpContext, name ApiName, chunk []byte, isLastChunk bool) ([]byte, error)
}

type StreamingEventHandler interface {
	OnStreamingEvent(ctx wrapper.HttpContext, name ApiName, event StreamEvent) ([]StreamEvent, error)
}

type ApiNameHandler interface {
	GetApiName(path string) ApiName
}

type TransformRequestHeadersHandler interface {
	TransformRequestHeaders(ctx wrapper.HttpContext, apiName ApiName, headers http.Header)
}

type TransformRequestBodyHandler interface {
	TransformRequestBody(ctx wrapper.HttpContext, apiName ApiName, body []byte) ([]byte, error)
}

// TransformRequestBodyHeadersHandler allows to transform request headers based on the request body.
// Some providers (e.g. gemini) transform request headers (e.g., path) based on the request body (e.g., model).
type TransformRequestBodyHeadersHandler interface {
	TransformRequestBodyHeaders(ctx wrapper.HttpContext, apiName ApiName, body []byte, headers http.Header) ([]byte, error)
}

type TransformResponseHeadersHandler interface {
	TransformResponseHeaders(ctx wrapper.HttpContext, apiName ApiName, headers http.Header)
}

type TransformResponseBodyHandler interface {
	TransformResponseBody(ctx wrapper.HttpContext, apiName ApiName, body []byte) ([]byte, error)
}

type ProviderConfig struct {
	// @Title zh-CN ID
	// @Description zh-CN AI服务提供商标识
	id string `required:"true" yaml:"id" json:"id"`
	// @Title zh-CN 类型
	// @Description zh-CN AI服务提供商类型
	typ string `required:"true" yaml:"type" json:"type"`
	// @Title zh-CN API Tokens
	// @Description zh-CN 在请求AI服务时用于认证的API Token列表。不同的AI服务提供商可能有不同的名称。部分供应商只支持配置一个API Token（如Azure OpenAI）。
	apiTokens []string `required:"false" yaml:"apiToken" json:"apiTokens"`
	// @Title zh-CN 请求超时
	// @Description zh-CN 请求AI服务的超时时间，单位为毫秒。默认值为120000，即2分钟。此项配置目前仅用于获取上下文信息，并不影响实际转发大模型请求。
	timeout uint32 `required:"false" yaml:"timeout" json:"timeout"`
	// @Title zh-CN apiToken 故障切换
	// @Description zh-CN 当 apiToken 不可用时移出 apiTokens 列表，对移除的 apiToken 进行健康检查，当重新可用后加回 apiTokens 列表
	failover *failover `required:"false" yaml:"failover" json:"failover"`
	// @Title zh-CN 失败请求重试
	// @Description zh-CN 对失败的请求立即进行重试
	retryOnFailure *retryOnFailure `required:"false" yaml:"retryOnFailure" json:"retryOnFailure"`
	// @Title zh-CN 推理内容处理方式
	// @Description zh-CN 如何处理大模型服务返回的推理内容。目前支持以下取值：passthrough（正常输出推理内容）、ignore（不输出推理内容）、concat（将推理内容拼接在常规输出内容之前）。默认为 normal。仅支持通义千问服务。
	reasoningContentMode string `required:"false" yaml:"reasoningContentMode" json:"reasoningContentMode"`
	// @Title zh-CN 基于OpenAI协议的自定义后端URL
	// @Description zh-CN 仅适用于支持 openai 协议的服务。
	openaiCustomUrl string `required:"false" yaml:"openaiCustomUrl" json:"openaiCustomUrl"`
	// @Title zh-CN Moonshot File ID
	// @Description zh-CN 仅适用于Moonshot AI服务。Moonshot AI服务的文件ID，其内容用于补充AI请求上下文
	moonshotFileId string `required:"false" yaml:"moonshotFileId" json:"moonshotFileId"`
	// @Title zh-CN Azure OpenAI Service URL
	// @Description zh-CN 仅适用于Azure OpenAI服务。要请求的OpenAI服务的完整URL，包含api-version等参数
	azureServiceUrl string `required:"false" yaml:"azureServiceUrl" json:"azureServiceUrl"`
	// @Title zh-CN 通义千问File ID
	// @Description zh-CN 仅适用于通义千问服务。上传到Dashscope的文件ID，其内容用于补充AI请求上下文。仅支持qwen-long模型。
	qwenFileIds []string `required:"false" yaml:"qwenFileIds" json:"qwenFileIds"`
	// @Title zh-CN 启用通义千问搜索服务
	// @Description zh-CN 仅适用于通义千问服务，表示是否启用通义千问的互联网搜索功能。
	qwenEnableSearch bool `required:"false" yaml:"qwenEnableSearch" json:"qwenEnableSearch"`
	// @Title zh-CN 通义千问服务域名
	// @Description zh-CN 仅适用于通义千问服务，默认转发域名为 dashscope.aliyuncs.com, 当使用金融云服务时，可以设置为 dashscope-finance.aliyuncs.com
	qwenDomain string `required:"false" yaml:"qwenDomain" json:"qwenDomain"`
	// @Title zh-CN 开启通义千问兼容模式
	// @Description zh-CN 启用通义千问兼容模式后，将调用千问的兼容模式接口，同时对请求/响应不做修改。
	qwenEnableCompatible bool `required:"false" yaml:"qwenEnableCompatible" json:"qwenEnableCompatible"`
	// @Title zh-CN Ollama Server IP/Domain
	// @Description zh-CN 仅适用于 Ollama 服务。Ollama 服务器的主机地址。
	ollamaServerHost string `required:"false" yaml:"ollamaServerHost" json:"ollamaServerHost"`
	// @Title zh-CN Ollama Server Port
	// @Description zh-CN 仅适用于 Ollama 服务。Ollama 服务器的端口号。
	ollamaServerPort uint32 `required:"false" yaml:"ollamaServerPort" json:"ollamaServerPort"`
	// @Title zh-CN hunyuan api key for authorization
	// @Description zh-CN 仅适用于Hun Yuan AI服务鉴权，API key/id 参考：https://cloud.tencent.com/document/api/1729/101843#Golang
	hunyuanAuthKey string `required:"false" yaml:"hunyuanAuthKey" json:"hunyuanAuthKey"`
	// @Title zh-CN hunyuan api id for authorization
	// @Description zh-CN 仅适用于Hun Yuan AI服务鉴权
	hunyuanAuthId string `required:"false" yaml:"hunyuanAuthId" json:"hunyuanAuthId"`
	// @Title zh-CN Amazon Bedrock AccessKey for authorization
	// @Description zh-CN 仅适用于Amazon Bedrock服务鉴权，API key/id 参考：https://docs.aws.amazon.com/zh_cn/IAM/latest/UserGuide/reference_sigv.html
	awsAccessKey string `required:"false" yaml:"awsAccessKey" json:"awsAccessKey"`
	// @Title zh-CN Amazon Bedrock SecretKey for authorization
	// @Description zh-CN 仅适用于Amazon Bedrock服务鉴权
	awsSecretKey string `required:"false" yaml:"awsSecretKey" json:"awsSecretKey"`
	// @Title zh-CN Amazon Bedrock Region
	// @Description zh-CN 仅适用于Amazon Bedrock服务访问
	awsRegion string `required:"false" yaml:"awsRegion" json:"awsRegion"`
	// @Title zh-CN Amazon Bedrock 额外模型请求参数
	// @Description zh-CN 仅适用于Amazon Bedrock服务，用于设置模型特定的推理参数
	bedrockAdditionalFields map[string]interface{} `required:"false" yaml:"bedrockAdditionalFields" json:"bedrockAdditionalFields"`
	// @Title zh-CN minimax API type
	// @Description zh-CN 仅适用于 minimax 服务。minimax API 类型，v2 和 pro 中选填一项，默认值为 v2
	minimaxApiType string `required:"false" yaml:"minimaxApiType" json:"minimaxApiType"`
	// @Title zh-CN minimax group id
	// @Description zh-CN 仅适用于 minimax 服务。minimax API 类型为 pro 时必填
	minimaxGroupId string `required:"false" yaml:"minimaxGroupId" json:"minimaxGroupId"`
	// @Title zh-CN 模型名称映射表
	// @Description zh-CN 用于将请求中的模型名称映射为目标AI服务商支持的模型名称。支持通过“*”来配置全局映射
	modelMapping map[string]string `required:"false" yaml:"modelMapping" json:"modelMapping"`
	// @Title zh-CN 对外接口协议
	// @Description zh-CN 通过本插件对外提供的AI服务接口协议。默认值为“openai”，即OpenAI的接口协议。如需保留原有接口协议，可配置为“original"
	protocol string `required:"false" yaml:"protocol" json:"protocol"`
	// @Title zh-CN 模型对话上下文
	// @Description zh-CN 配置一个外部获取对话上下文的文件来源，用于在AI请求中补充对话上下文
	context *ContextConfig `required:"false" yaml:"context" json:"context"`
	// @Title zh-CN 版本
	// @Description zh-CN 请求AI服务的版本，目前仅适用于 Gemini 和 Claude AI服务
	apiVersion string `required:"false" yaml:"apiVersion" json:"apiVersion"`
	// @Title zh-CN Cloudflare Account ID
	// @Description zh-CN 仅适用于 Cloudflare Workers AI 服务。参考：https://developers.cloudflare.com/workers-ai/get-started/rest-api/#2-run-a-model-via-api
	cloudflareAccountId string `required:"false" yaml:"cloudflareAccountId" json:"cloudflareAccountId"`
	// @Title zh-CN Gemini AI内容过滤和安全级别设定
	// @Description zh-CN 仅适用于 Gemini AI 服务。参考：https://ai.google.dev/gemini-api/docs/safety-settings
	geminiSafetySetting map[string]string `required:"false" yaml:"geminiSafetySetting" json:"geminiSafetySetting"`
	// @Title zh-CN Vertex AI访问区域
	// @Description zh-CN 仅适用于Vertex AI服务。如需查看支持的区域的完整列表，请参阅https://cloud.google.com/vertex-ai/generative-ai/docs/learn/locations?hl=zh-cn#available-regions
	vertexRegion string `required:"false" yaml:"vertexRegion" json:"vertexRegion"`
	// @Title zh-CN Vertex AI项目Id
	// @Description zh-CN 仅适用于Vertex AI服务。创建和管理项目请参阅https://cloud.google.com/resource-manager/docs/creating-managing-projects?hl=zh-cn#identifiers
	vertexProjectId string `required:"false" yaml:"vertexProjectId" json:"vertexProjectId"`
	// @Title zh-CN Vertex 认证秘钥
	// @Description zh-CN 用于Google服务账号认证的完整JSON密钥文件内容，获取可参考https://cloud.google.com/iam/docs/keys-create-delete?hl=zh-cn#iam-service-account-keys-create-console
	vertexAuthKey string `required:"false" yaml:"vertexAuthKey" json:"vertexAuthKey"`
	// @Title zh-CN Vertex 认证服务名
	// @Description zh-CN 用于Google服务账号认证的服务,DNS类型的服务名
	vertexAuthServiceName string `required:"false" yaml:"vertexAuthServiceName" json:"vertexAuthServiceName"`
	// @Title zh-CN Vertex token刷新提前时间
	// @Description zh-CN 用于Google服务账号认证，access token过期时间判定提前刷新，单位为秒，默认值为60秒
	vertexTokenRefreshAhead int64 `required:"false" yaml:"vertexTokenRefreshAhead" json:"vertexTokenRefreshAhead"`
	// @Title zh-CN 翻译服务需指定的目标语种
	// @Description zh-CN 翻译结果的语种，目前仅适用于DeepL服务。
	targetLang string `required:"false" yaml:"targetLang" json:"targetLang"`
	// @Title zh-CN  指定服务返回的响应需满足的JSON Schema
	// @Description zh-CN 目前仅适用于OpenAI部分模型服务。参考：https://platform.openai.com/docs/guides/structured-outputs
	responseJsonSchema map[string]interface{} `required:"false" yaml:"responseJsonSchema" json:"responseJsonSchema"`
	// @Title zh-CN 自定义大模型参数配置
	// @Description zh-CN 用于填充或者覆盖大模型调用时的参数
	customSettings []CustomSetting
	// @Title zh-CN dify私有化部署的url
	difyApiUrl string `required:"false" yaml:"difyApiUrl" json:"difyApiUrl"`
	// @Title zh-CN dify的应用类型，Chat/Completion/Agent/Workflow
	botType string `required:"false" yaml:"botType" json:"botType"`
	// @Title zh-CN dify中应用类型为workflow时需要设置输入变量，当botType为workflow时一起使用
	inputVariable string `required:"false" yaml:"inputVariable" json:"inputVariable"`
	// @Title zh-CN dify中应用类型为workflow时需要设置输出变量，当botType为workflow时一起使用
	outputVariable string `required:"false" yaml:"outputVariable" json:"outputVariable"`
	// @Title zh-CN 额外支持的ai能力
	// @Description zh-CN 开放的ai能力和urlpath映射，例如： {"openai/v1/chatcompletions": "/v1/chat/completions"}
	capabilities map[string]string
	// @Title zh-CN 如果配置了basePath，可用于在请求path中移除该前缀，或添加至请求path中，默认为进行移除
	basePath string `required:"false" yaml:"basePath" json:"basePath"`
	// @Title zh-CN basePathHandling用于指定basePath的处理方式，可选值：removePrefix、prepend
	basePathHandling basePathHandling `required:"false" yaml:"basePathHandling" json:"basePathHandling"`
	// @Title zh-CN 首包超时
	// @Description zh-CN 流式请求中收到上游服务第一个响应包的超时时间，单位为毫秒。默认值为 0，表示不开启首包超时
	firstByteTimeout uint32 `required:"false" yaml:"firstByteTimeout" json:"firstByteTimeout"`
}

func (c *ProviderConfig) GetId() string {
	return c.id
}

func (c *ProviderConfig) GetType() string {
	return c.typ
}

func (c *ProviderConfig) GetProtocol() string {
	return c.protocol
}

func (c *ProviderConfig) IsOpenAIProtocol() bool {
	return c.protocol == protocolOpenAI
}

func (c *ProviderConfig) FromJson(json gjson.Result) {
	c.id = json.Get("id").String()
	c.typ = json.Get("type").String()
	c.apiTokens = make([]string, 0)
	for _, token := range json.Get("apiTokens").Array() {
		c.apiTokens = append(c.apiTokens, token.String())
	}
	c.timeout = uint32(json.Get("timeout").Uint())
	if c.timeout == 0 {
		c.timeout = defaultTimeout
	}
	// first byte timeout
	c.firstByteTimeout = uint32(json.Get("firstByteTimeout").Uint())
	c.openaiCustomUrl = json.Get("openaiCustomUrl").String()
	c.moonshotFileId = json.Get("moonshotFileId").String()
	c.azureServiceUrl = json.Get("azureServiceUrl").String()
	c.qwenFileIds = make([]string, 0)
	for _, fileId := range json.Get("qwenFileIds").Array() {
		c.qwenFileIds = append(c.qwenFileIds, fileId.String())
	}
	c.qwenEnableSearch = json.Get("qwenEnableSearch").Bool()
	c.qwenEnableCompatible = json.Get("qwenEnableCompatible").Bool()
	c.qwenDomain = json.Get("qwenDomain").String()
	if c.qwenDomain != "" {
		// TODO: validate the domain, if not valid, set to default
	}
	c.ollamaServerHost = json.Get("ollamaServerHost").String()
	c.ollamaServerPort = uint32(json.Get("ollamaServerPort").Uint())
	c.modelMapping = make(map[string]string)
	for k, v := range json.Get("modelMapping").Map() {
		c.modelMapping[k] = v.String()
	}
	c.protocol = json.Get("protocol").String()
	if c.protocol == "" {
		c.protocol = protocolOpenAI
	}
	contextJson := json.Get("context")
	if contextJson.Exists() {
		c.context = &ContextConfig{}
		c.context.FromJson(contextJson)
	}

	// 这里获取 claudeVersion 字段，与结构体中定义 yaml/json 的 tag 不一致
	c.apiVersion = json.Get("claudeVersion").String()
	if c.apiVersion == "" {
		// 增加获取 version 字段，用于适配其他模型的配置，并保持与结构体中定义的 tag 一致
		c.apiVersion = json.Get("apiVersion").String()
	}
	c.hunyuanAuthId = json.Get("hunyuanAuthId").String()
	c.hunyuanAuthKey = json.Get("hunyuanAuthKey").String()
	c.awsAccessKey = json.Get("awsAccessKey").String()
	c.awsSecretKey = json.Get("awsSecretKey").String()
	c.awsRegion = json.Get("awsRegion").String()
	if c.typ == providerTypeBedrock {
		c.bedrockAdditionalFields = make(map[string]interface{})
		for k, v := range json.Get("bedrockAdditionalFields").Map() {
			c.bedrockAdditionalFields[k] = v.Value()
		}
	}
	c.minimaxApiType = json.Get("minimaxApiType").String()
	c.minimaxGroupId = json.Get("minimaxGroupId").String()
	c.cloudflareAccountId = json.Get("cloudflareAccountId").String()
	if c.typ == providerTypeGemini || c.typ == providerTypeVertex {
		c.geminiSafetySetting = make(map[string]string)
		for k, v := range json.Get("geminiSafetySetting").Map() {
			c.geminiSafetySetting[k] = v.String()
		}
	}
	c.vertexRegion = json.Get("vertexRegion").String()
	c.vertexProjectId = json.Get("vertexProjectId").String()
	c.vertexAuthKey = json.Get("vertexAuthKey").String()
	c.vertexAuthServiceName = json.Get("vertexAuthServiceName").String()
	c.vertexTokenRefreshAhead = json.Get("vertexTokenRefreshAhead").Int()
	if c.vertexTokenRefreshAhead == 0 {
		c.vertexTokenRefreshAhead = 60
	}
	c.targetLang = json.Get("targetLang").String()

	if schemaValue, ok := json.Get("responseJsonSchema").Value().(map[string]interface{}); ok {
		c.responseJsonSchema = schemaValue
	} else {
		c.responseJsonSchema = nil
	}

	c.customSettings = make([]CustomSetting, 0)
	customSettingsJson := json.Get("customSettings")
	if customSettingsJson.Exists() {
		protocol := protocolOpenAI
		if c.protocol == protocolOriginal {
			// use provider name to represent original protocol name
			protocol = c.typ
		}
		for _, settingJson := range customSettingsJson.Array() {
			setting := CustomSetting{}
			setting.FromJson(settingJson)
			// use protocol info to rewrite setting
			setting.AdjustWithProtocol(protocol)
			if setting.Validate() {
				c.customSettings = append(c.customSettings, setting)
			}
		}
	}

	c.reasoningContentMode = json.Get("reasoningContentMode").String()
	if c.reasoningContentMode == "" {
		c.reasoningContentMode = reasoningBehaviorPassThrough
	} else {
		c.reasoningContentMode = strings.ToLower(c.reasoningContentMode)
		switch c.reasoningContentMode {
		case reasoningBehaviorPassThrough, reasoningBehaviorIgnore, reasoningBehaviorConcat:
			break
		default:
			c.reasoningContentMode = reasoningBehaviorPassThrough
			break
		}
	}

	failoverJson := json.Get("failover")
	c.failover = &failover{
		enabled: false,
	}
	if failoverJson.Exists() {
		c.failover.FromJson(failoverJson)
	}

	retryOnFailureJson := json.Get("retryOnFailure")
	c.retryOnFailure = &retryOnFailure{
		enabled: false,
	}
	if retryOnFailureJson.Exists() {
		c.retryOnFailure.FromJson(retryOnFailureJson)
	}
	c.difyApiUrl = json.Get("difyApiUrl").String()
	c.botType = json.Get("botType").String()
	c.inputVariable = json.Get("inputVariable").String()
	c.outputVariable = json.Get("outputVariable").String()

	c.capabilities = make(map[string]string)
	for capability, pathJson := range json.Get("capabilities").Map() {
		// 过滤掉不受支持的能力
		switch capability {
		case string(ApiNameChatCompletion),
			string(ApiNameEmbeddings),
			string(ApiNameImageGeneration),
			string(ApiNameImageVariation),
			string(ApiNameImageEdit),
			string(ApiNameAudioSpeech),
			string(ApiNameCohereV1Rerank):
			c.capabilities[capability] = pathJson.String()
		}
	}
	c.basePath = json.Get("basePath").String()
	c.basePathHandling = basePathHandling(json.Get("basePathHandling").String())
	if c.basePath != "" && c.basePathHandling == "" {
		c.basePathHandling = basePathHandlingRemovePrefix
	}
}

func (c *ProviderConfig) Validate() error {
	if c.protocol != protocolOpenAI && c.protocol != protocolOriginal {
		return errors.New("invalid protocol in config")
	}
	if c.context != nil {
		if err := c.context.Validate(); err != nil {
			return err
		}
	}

	if c.failover.enabled {
		if err := c.failover.Validate(); err != nil {
			return err
		}
	}

	if c.typ == "" {
		return errors.New("missing type in provider config")
	}
	initializer, has := providerInitializers[c.typ]
	if !has {
		return errors.New("unknown provider type: " + c.typ)
	}
	if err := initializer.ValidateConfig(c); err != nil {
		return err
	}
	return nil
}

func (c *ProviderConfig) GetOrSetTokenWithContext(ctx wrapper.HttpContext) string {
	ctxApiKey := ctx.GetContext(ctxKeyApiKey)
	if ctxApiKey == nil {
		ctxApiKey = c.GetRandomToken()
		ctx.SetContext(ctxKeyApiKey, ctxApiKey)
	}
	return ctxApiKey.(string)
}

func (c *ProviderConfig) GetRandomToken() string {
	apiTokens := c.apiTokens
	count := len(apiTokens)
	switch count {
	case 0:
		return ""
	case 1:
		return apiTokens[0]
	default:
		return apiTokens[rand.Intn(count)]
	}
}

func (c *ProviderConfig) IsOriginal() bool {
	return c.protocol == protocolOriginal
}

func (c *ProviderConfig) ReplaceByCustomSettings(body []byte) ([]byte, error) {
	return ReplaceByCustomSettings(body, c.customSettings)
}

func CreateProvider(pc ProviderConfig) (Provider, error) {
	initializer, has := providerInitializers[pc.typ]
	if !has {
		return nil, errors.New("unknown provider type: " + pc.typ)
	}
	return initializer.CreateProvider(pc)
}

func (c *ProviderConfig) parseRequestAndMapModel(ctx wrapper.HttpContext, request interface{}, body []byte) error {
	switch req := request.(type) {
	case *chatCompletionRequest:
		if err := decodeChatCompletionRequest(body, req); err != nil {
			return err
		}

		streaming := req.Stream
		if streaming {
			_ = proxywasm.ReplaceHttpRequestHeader("Accept", "text/event-stream")
			ctx.SetContext(ctxKeyIsStreaming, true)
		} else {
			ctx.SetContext(ctxKeyIsStreaming, false)
		}

		return c.setRequestModel(ctx, req)
	case *embeddingsRequest:
		if err := decodeEmbeddingsRequest(body, req); err != nil {
			return err
		}
		return c.setRequestModel(ctx, req)
	case *imageGenerationRequest:
		if err := decodeImageGenerationRequest(body, req); err != nil {
			return err
		}
		return c.setRequestModel(ctx, req)
	default:
		return errors.New("unsupported request type")
	}
}

func (c *ProviderConfig) setRequestModel(ctx wrapper.HttpContext, request interface{}) error {
	var model *string

	switch req := request.(type) {
	case *chatCompletionRequest:
		model = &req.Model
	case *embeddingsRequest:
		model = &req.Model
	case *imageGenerationRequest:
		model = &req.Model
	default:
		return errors.New("unsupported request type")
	}

	return c.mapModel(ctx, model)
}

func (c *ProviderConfig) mapModel(ctx wrapper.HttpContext, model *string) error {
	if *model == "" {
		return errors.New("missing model in request")
	}
	ctx.SetContext(ctxKeyOriginalRequestModel, *model)

	mappedModel := getMappedModel(*model, c.modelMapping)
	if mappedModel == "" {
		return errors.New("model becomes empty after applying the configured mapping")
	}

	*model = mappedModel
	ctx.SetContext(ctxKeyFinalRequestModel, *model)
	return nil
}

func getMappedModel(model string, modelMapping map[string]string) string {
	mappedModel := doGetMappedModel(model, modelMapping)
	if len(mappedModel) != 0 {
		return mappedModel
	}
	return model
}

func doGetMappedModel(model string, modelMapping map[string]string) string {
	if len(modelMapping) == 0 {
		return ""
	}

	if v, ok := modelMapping[model]; ok {
		log.Debugf("model [%s] is mapped to [%s] explictly", model, v)
		return v
	}

	for k, v := range modelMapping {
		if k == wildcard {
			continue
		}
		if strings.HasSuffix(k, wildcard) {
			k = strings.TrimSuffix(k, wildcard)
			if strings.HasPrefix(model, k) {
				log.Debugf("model [%s] is mapped to [%s] via prefix [%s]", model, v, k)
				return v
			}
		}

		if strings.HasPrefix(k, "~") {
			k = strings.TrimPrefix(k, "~")
			re := regexp.MustCompile(k)
			if re.MatchString(model) {
				v = re.ReplaceAllString(model, v)
				log.Debugf("model [%s] is mapped to [%s] via regex [%s]", model, v, k)
				return v
			}
		}
	}

	if v, ok := modelMapping[wildcard]; ok {
		log.Debugf("model [%s] is mapped to [%s] via wildcard", model, v)
		return v
	}

	return ""
}

func ExtractStreamingEvents(ctx wrapper.HttpContext, chunk []byte) []StreamEvent {
	body := chunk
	if bufferedStreamingBody, has := ctx.GetContext(ctxKeyStreamingBody).([]byte); has {
		body = append(bufferedStreamingBody, chunk...)
	}
	body = bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
	body = bytes.ReplaceAll(body, []byte("\r"), []byte("\n"))

	eventStartIndex, lineStartIndex, valueStartIndex := -1, -1, -1

	defer func() {
		if eventStartIndex >= 0 && eventStartIndex < len(body) {
			// Just in case the received chunk is not a complete event.
			ctx.SetContext(ctxKeyStreamingBody, body[eventStartIndex:])
		} else {
			ctx.SetContext(ctxKeyStreamingBody, nil)
		}
	}()

	// Sample Qwen event response:
	//
	// event:result
	// :HTTP_STATUS/200
	// data:{"output":{"choices":[{"message":{"content":"你好！","role":"assistant"},"finish_reason":"null"}]},"usage":{"total_tokens":116,"input_tokens":114,"output_tokens":2},"request_id":"71689cfc-1f42-9949-86e8-9563b7f832b1"}
	//
	// event:error
	// :HTTP_STATUS/400
	// data:{"code":"InvalidParameter","message":"Preprocessor error","request_id":"0cbe6006-faec-9854-bf8b-c906d75c3bd8"}
	//

	var events []StreamEvent

	currentKey := ""
	currentEvent := &StreamEvent{}
	i, length := 0, len(body)
	for i = 0; i < length; i++ {
		ch := body[i]
		if ch != '\n' {
			if lineStartIndex == -1 {
				if eventStartIndex == -1 {
					eventStartIndex = i
				}
				lineStartIndex = i
				valueStartIndex = -1
			}
			if valueStartIndex == -1 {
				if ch == ':' {
					valueStartIndex = i + 1
					currentKey = string(body[lineStartIndex:valueStartIndex])
				}
			} else if valueStartIndex == i && ch == ' ' {
				// Skip leading spaces in data.
				valueStartIndex = i + 1
			}
			continue
		}

		if lineStartIndex != -1 {
			value := string(body[valueStartIndex:i])
			currentEvent.SetValue(currentKey, value)
		} else {
			// Extra new line. The current event is complete.
			events = append(events, *currentEvent)
			// Reset event parsing state.
			eventStartIndex = -1
			currentEvent = &StreamEvent{}
		}

		// Reset line parsing state.
		lineStartIndex = -1
		valueStartIndex = -1
		currentKey = ""
	}

	return events
}

func (c *ProviderConfig) isSupportedAPI(apiName ApiName) bool {
	_, exist := c.capabilities[string(apiName)]
	return exist
}

func (c *ProviderConfig) setDefaultCapabilities(capabilities map[string]string) {
	for capability, path := range capabilities {
		c.capabilities[capability] = path
	}
}

func (c *ProviderConfig) handleRequestBody(
	provider Provider, contextCache *contextCache, ctx wrapper.HttpContext, apiName ApiName, body []byte,
) (types.Action, error) {
	// add the first byte timeout header to the request
	if c.firstByteTimeout != 0 && c.isStreamingAPI(apiName, body) {
		err := proxywasm.ReplaceHttpRequestHeader("x-envoy-upstream-rq-first-byte-timeout-ms", strconv.FormatUint(uint64(c.firstByteTimeout), 10))
		if err != nil {
			log.Errorf("failed to set x-envoy-upstream-rq-first-byte-timeout-ms header: %v", err)
		}
		log.Debugf("[firstByteTimeout] %d", c.firstByteTimeout)
	}

	// use original protocol
	if c.IsOriginal() {
		return types.ActionContinue, nil
	}

	// use openai protocol
	var err error
	if handler, ok := provider.(TransformRequestBodyHandler); ok {
		body, err = handler.TransformRequestBody(ctx, apiName, body)
	} else if handler, ok := provider.(TransformRequestBodyHeadersHandler); ok {
		headers := util.GetOriginalRequestHeaders()
		body, err = handler.TransformRequestBodyHeaders(ctx, apiName, body, headers)
		util.ReplaceRequestHeaders(headers)
	} else {
		body, err = c.defaultTransformRequestBody(ctx, apiName, body)
	}

	if err != nil {
		return types.ActionContinue, err
	}

	if apiName == ApiNameChatCompletion {
		if c.context == nil {
			return types.ActionContinue, replaceRequestBody(body)
		}
		err = contextCache.GetContextFromFile(ctx, provider, body)

		if err == nil {
			return types.ActionPause, nil
		}
		return types.ActionContinue, err
	}
	return types.ActionContinue, replaceRequestBody(body)
}

func (c *ProviderConfig) handleRequestHeaders(provider Provider, ctx wrapper.HttpContext, apiName ApiName) {
	headers := util.GetOriginalRequestHeaders()
	originPath := headers.Get(":path")
	if c.basePath != "" && c.basePathHandling == basePathHandlingRemovePrefix {
		headers.Set(":path", strings.TrimPrefix(originPath, c.basePath))
	}
	if handler, ok := provider.(TransformRequestHeadersHandler); ok {
		handler.TransformRequestHeaders(ctx, apiName, headers)
	}
	if c.basePath != "" && c.basePathHandling == basePathHandlingPrepend && !strings.HasPrefix(headers.Get(":path"), c.basePath) {
		headers.Set(":path", path.Join(c.basePath, headers.Get(":path")))
	}
	if headers.Get(":path") != originPath {
		headers.Set("X-ENVOY-ORIGINAL-PATH", originPath)
	}
	util.ReplaceRequestHeaders(headers)
}

// defaultTransformRequestBody 默认的请求体转换方法，只做模型映射，用slog替换模型名称，不用序列化和反序列化，提高性能
func (c *ProviderConfig) defaultTransformRequestBody(ctx wrapper.HttpContext, apiName ApiName, body []byte) ([]byte, error) {
	switch apiName {
	case ApiNameChatCompletion:
		stream := gjson.GetBytes(body, "stream").Bool()
		if stream {
			_ = proxywasm.ReplaceHttpRequestHeader("Accept", "text/event-stream")
			ctx.SetContext(ctxKeyIsStreaming, true)
		} else {
			ctx.SetContext(ctxKeyIsStreaming, false)
		}
	}
	model := gjson.GetBytes(body, "model").String()
	ctx.SetContext(ctxKeyOriginalRequestModel, model)
	return sjson.SetBytes(body, "model", getMappedModel(model, c.modelMapping))
}

func (c *ProviderConfig) DefaultTransformResponseHeaders(ctx wrapper.HttpContext, headers http.Header) {
	if c.protocol == protocolOriginal {
		ctx.DontReadResponseBody()
	} else {
		headers.Del("Content-Length")
	}
}

func (c *ProviderConfig) isStreamingAPI(apiName ApiName, body []byte) bool {
	stream := false
	switch apiName {
	case ApiNameCompletion,
		ApiNameChatCompletion,
		ApiNameImageGeneration,
		ApiNameImageEdit,
		ApiNameResponses,
		ApiNameQwenAsyncAIGC,
		ApiNameAnthropicMessages,
		ApiNameAnthropicComplete:
		stream = gjson.GetBytes(body, "stream").Bool()
	case ApiNameGeminiStreamGenerateContent:
		stream = true
	}
	return stream
}

func (c *ProviderConfig) needToProcessRequestBody(apiName ApiName) bool {
	switch apiName {
	case ApiNameChatCompletion,
		ApiNameCompletion,
		ApiNameEmbeddings,
		ApiNameImageGeneration,
		ApiNameImageEdit,
		ApiNameImageVariation,
		ApiNameAudioSpeech,
		ApiNameFineTuningJobs,
		ApiNameResponses,
		ApiNameGeminiGenerateContent,
		ApiNameGeminiStreamGenerateContent,
		ApiNameAnthropicMessages:
		return true
	}
	return false
}
