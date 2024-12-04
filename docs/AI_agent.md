# AI Agent

```go
func run() error {
 llm, err := ollama.New(ollama.WithModel("llama3.1"))
 if err != nil {
  fmt.Println(err)
  return err
 }
 search, err := serpapi.New()
 if err != nil {
  return err
 }
 agentTools := []tools.Tool{
  tools.Calculator{},
  search,
 }

 agent := agents.NewOneShotAgent(llm,
  agentTools,
  agents.WithMaxIterations(3))
 executor := agents.NewExecutor(agent)

 question := "Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?"
 answer, err := chains.Run(context.Background(), executor, question)
 fmt.Println(answer)
 return err
}
```

-- agent 把llm封装未一个 chain

```text
Chain.run
callChain->
 Executor.Call->
    doIteration->
        e.Agent.Plan(agent)->
            chains.Predict->
                call->
                    llmchain->
                        call
                            c.Prompt.FormatPrompt(values) 格式化prompt
                            llms.GenerateFromSinglePrompt
                            llm.GenerateContent 真是和AI发送http请求
                        c.OutputParser.ParseWithPrompt
            parseOutput 返回 schema.AgentAction
                解析结果生成tool列连 调用tool.call
    e.doAction
        tool.Call
```

## 交互

### LLM

- input

    ```markdown
    Today is November 27, 2024.
    Answer the following questions as best you can. You have access to the following tools:

    - calculator: Useful for getting the result of a math expression.
    The input to this tool should be a valid mathematical expression that could be executed by a starlark evaluator.
    - GoogleSearch:
    "A wrapper around Google Search. "
    "Useful for when you need to answer questions about current events. "
    "Always one of the first options when you need to find information on internet"
    "Input should be a search query."


    Use the following format:

    Question: the input question you must answer
    Thought: you should always think about what to do
    Action: the action to take, should be one of [ calculator, GoogleSearch ]
    Action Input: the input to the action
    Observation: the result of the action
    ... (this Thought/Action/Action Input/Observation can repeat N times)
    Thought: I now know the final answer
    Final Answer: the final answer to the original input question

    Begin!

    Question: Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?
    ```

- Action

    ```markdown
    Thought: To find out who Olivia Wilde's boyfriend is, I should search for her current relationship status.
    Action: GoogleSearch
    Action Input: "Olivia Wilde boyfriend"
    ```

### tool google search

- Input

    ```text
    "Olivia Wilde boyfriend"
    ```

- Output

    ```text
    26 November 2024... Olivia Wilde news, gossip, photos of Olivia Wilde, biography, Olivia Wilde boyfriend list 2024. Relationship history.
    ```

### 第二次 LLM

- Input

    ```text
    第一次 prompt
    Thought: To find out who Olivia Wilde's boyfriend is, I should search for her current relationship status.
    Action: GoogleSearch
    Action Input: "Olivia Wilde boyfriend"
    Observation: 26 November 2024... Olivia Wilde news, gossip, photos of Olivia Wilde, biography, Olivia Wilde boyfriend list 2024. Relationship history.
    ```

- Output

    ```text
    Thought: It looks like the search result is from today's date, which matches our current year.
    Action: GoogleSearch
    Action Input: "Olivia Wilde boyfriend November 27, 2024"
    ```

### 第二次 tool google search

- Input

    ```text
    "Olivia Wilde boyfriend November 27, 2024"
    ```

- Output

    ```text
    Olivia Wilde and Jason Sudeikis might not be together anymore, but they're still friendly as co-parents. The former couple was spotted ...
    ```

### 第三次 LLM

- Input

    ```text
    前两次 prompt
    Thought: It looks like the search result is from today's date, which matches our current year.
    Action: GoogleSearch
    Action Input: "Olivia Wilde boyfriend November 27, 2024"
    Observation: Olivia Wilde and Jason Sudeikis might not be together anymore, but they're still friendly as co-parents. The former couple was spotted ...
    ```

- Output

    ```text
    Thought: I now know that Olivia Wilde is no longer with Jason Sudeikis.
    Action: GoogleSearch
    Action Input: "Jason Sudeikis current age"
    ```
