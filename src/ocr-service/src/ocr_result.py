class OCRResult:
    def __init__(self, text_result: str, meeting_id: str, id=None):
        self.text_result = text_result
        self.meeting_id = meeting_id
        self.id = id

    def to_dict(self):
        result = {
            "text_result": self.text_result,
            "meeting_id": self.meeting_id,
        }
        if self.id is not None:
            result["_id"] = self.id
        return result

    @staticmethod
    def from_dict(data: dict):
        return OCRResult(
            id=data.get("_id"),
            text_result=data.get("text_result", ""),
            meeting_id=data.get("meeting_id", ""),
        )
