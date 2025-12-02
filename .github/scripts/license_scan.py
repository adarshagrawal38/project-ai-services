import argparse
import json
import sys
import csv

def parse_cyclonedx(data, tool_name):
    sbom = {}
    if "components" in data:
        for component in data["components"]:
            try:
                licenses = ', '.join(
                    map(
                        lambda lic: lic["expression"] if "expression" in lic else (
                            lic["license"].get("id") if isinstance(lic.get("license"), dict) and "id" in lic["license"] else lic["license"]["name"]
                        ),
                        component.get("licenses", [])
                    )
                )
            except KeyError:
                licenses = "-"
            name = component.get("name", "-")
            version = component.get("version", "-")
            if name != "-" and version != "-":
                sbom[f"{name}@{version}"] = {f"License by {tool_name}": licenses}
    return sbom

def merge_sbom_data(trivy_data, parlay_data):
    merged_data = {}
    for key, value in trivy_data.items():
        merged_data[key] = value
        merged_data[key].setdefault("License by Parlay", "-")
    for key, value in parlay_data.items():
        if key in merged_data:
            merged_data[key].update(value)
        else:
            merged_data[key] = value
            merged_data[key].setdefault("License by Trivy", "-")
    return merged_data

def get_license_list(trivy_data):
    deny_pkg = {}
    warn_pkg = {}
    allowed_pkg = {}
    
    for key, value in trivy_data.items():
        license_value = value["License by Trivy"] + " " + value["License by Parlay"]
        if check_deny(license_value):
            deny_pkg[key] = value
        elif check_gpl_license(license_value):
            warn_pkg[key] = value
        else:
            allowed_pkg[key] = value
    return deny_pkg, warn_pkg, allowed_pkg

def check_deny(license_name):
    deny_list = ["AGPL-3.0", "SSPL-1.0", "RPL-1.5", "OSL-3.0", "CPAL-1.0"]
    for deny in deny_list:
        if deny in license_name:
            return True
    return False

def check_gpl_license(license_name):
    gpl_license_list = ["GPL-1.0", "GPL-2.0", "GPL-3.0"]
    for val in gpl_license_list:
        if val in license_name:
            return True
    return False

def print_result(licence_list, msg):
    if len(licence_list) == 0:
        return
    print()
    print(msg)
    print("{:<50} | {:<30} | {:<30} | {:<30}".format("Name", "Version", "License by Trivy", "License by Parlay"))
    print(f"{'-'*50} + {'-'*30} + {'-'*30} + {'-'*30}")
    for key, value in licence_list.items():
        name, version = key.split("@")
        print("{:<50} | {:<30} | {:<30} | {:<30}".format(name, version, value.get("License by Trivy", "-"), value.get("License by Parlay", "-")))
    print()

def generate_final_license_list(merged_data):
    with open("Final_license_list.csv", "w", newline="") as csv_file:
        writer = csv.writer(csv_file, delimiter="|")
        writer.writerow(["Package Name", "Version", "License by Trivy", "License by Parlay"])
        for key, value in merged_data.items():
            name, version = key.split("@")
            writer.writerow([name, version, value.get("License by Trivy", "-"), value.get("License by Parlay", "-")])

if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog='SBOM Gen', description='Generate readable SBOM combining Trivy and Parlay data')
    parser.add_argument('trivy_file', help="File path for Trivy CycloneDX input", default="sbom-trivy.json")
    parser.add_argument('parlay_file', help="File path for Parlay CycloneDX input")
    args = parser.parse_args()
    
    with open(args.trivy_file, encoding="utf-8") as trivy_file:
        trivy_data = parse_cyclonedx(json.load(trivy_file), "Trivy")
    
    with open(args.parlay_file, encoding="utf-8") as parlay_file:
        parlay_data = parse_cyclonedx(json.load(parlay_file), "Parlay")
        
    merged_data = merge_sbom_data(trivy_data, parlay_data)
    
    deny_list, warn_list, allowed_list = get_license_list(merged_data)
    
    error_exit = False
    generate_final_license_list(merged_data)
    
    print_result(warn_list, f"\n\n{'*'*40} WARNINGS {'*'*40}\n\n")
    print_result(deny_list, f"\n\n{'*'*40} DENIED {'*'*40}\n\n")
    
    if len(deny_list) > 0:
        print("Error: Dected packages with denied licenses, please removed the package with denied licenses")
        sys.exit(1)
