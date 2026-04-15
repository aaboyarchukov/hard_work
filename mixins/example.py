class XMLParser:
	def parse_to_xml(self, content : list[bytes]) -> XMLFile:
		...
		
class JSONParser:
	def parse_to_json(self, content : list[bytes]) -> JSONFile:
		...
		
class WebClient(XMLParser, JSONParser):
	def __init__(self, content):
		self.content = content
		
client = WebClient(list[bytes][0, 1, 0, 1, 1, 1])
xml = client.parse_to_xml(client.content)
json = client.parse_to_json(client.content)