import os

from dotenv import load_dotenv
from pydantic_ai import Agent
from pydantic_ai.models.openai import OpenAIChatModel
from pydantic_ai.providers.openai import OpenAIProvider

load_dotenv()

model = OpenAIChatModel(
    model_name="gpt-5.6",
    provider=OpenAIProvider(api_key=os.getenv("OPENAI_API_KEY")),
)

agent = Agent(model=model, retries=5)
result_sync = agent.run_sync("напиши сортировку пеерстановкой на python")
print(result_sync.output)
