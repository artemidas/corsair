import { Dialogs } from "@wailsio/runtime";

const YAML_FILTERS = [{ DisplayName: "YAML", Pattern: "*.yaml;*.yml" }];

export async function confirmDelete(
  message: string,
  title = "Delete",
): Promise<boolean> {
  const answer = await Dialogs.Question({
    Title: title,
    Message: message,
    Buttons: [
      { Label: "Cancel", IsCancel: true },
      { Label: "Delete", IsDefault: true },
    ],
  });
  return answer === "Delete";
}

export async function openYamlFile(title: string): Promise<string | null> {
  const selected = await Dialogs.OpenFile({
    Title: title,
    CanChooseFiles: true,
    AllowsMultipleSelection: false,
    Filters: YAML_FILTERS,
  });
  if (!selected || Array.isArray(selected)) return null;
  return selected;
}

export async function saveYamlFile(
  title: string,
  filename: string,
): Promise<string | null> {
  const dest = await Dialogs.SaveFile({
    Title: title,
    Filename: filename,
    Filters: YAML_FILTERS,
  });
  return dest || null;
}
