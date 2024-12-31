import json
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from helpers.auth import client


def main():
    sf = client()
    desc = sf.Account.describe()
    print("Available fields for Account objects:")
    for field in desc["fields"]:
        if field["name"] not in [
            "Id",
            "IsDeleted",
            "MasterRecordId",
            "ParentId",
            "BillingLatitude",
            "BillingLongitude",
            "BillingGeocodeAccuracy",
            "ShippingLatitude",
            "ShippingLongitude",
            "ShippingGeocodeAccuracy",
            "Sic",
            "OwnerId",
            "SystemModstamp",
            "LastReferencedDate",
            "Jigsaw",
            "JigsawCompanyId",
            "Tradestyle",
            "SicDesc",
        ]:
            print(
                f"name: {field['name']} - label: {field['label']} - type: {field['type']} - picklistValues: {json.dumps(field.get('picklistValues'))}"
            )


if __name__ == "__main__":
    main()