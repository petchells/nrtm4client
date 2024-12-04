import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Stack from "@mui/material/Stack";

export default function MenuContent(props: {
  mainItems: any[];
  secondaryItems: any[];
  menuItemSelectedIdx: number;
  onSelected: (idx: number) => void;
}) {
  // handle click on menu item
  const menuClicked = (idx: number) => () => {
    console.log("selected", idx);
    props.onSelected(idx);
    // setMenuItemSelectedIdx(idx);
    // if (mainListItems[idx].href) {
    //   window.location.href = mainListItems[idx].href;
    // }
  };

  return (
    <Stack sx={{ flexGrow: 1, p: 1, justifyContent: "space-between" }}>
      <List dense>
        {props.mainItems.map((item, index) => (
          <ListItem key={index} disablePadding sx={{ display: "block" }}>
            <ListItemButton
              selected={index === props.menuItemSelectedIdx}
              onClick={menuClicked(index)}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
      <List dense>
        {props.secondaryItems.map((item, index) => (
          <ListItem key={index} disablePadding sx={{ display: "block" }}>
            <ListItemButton>
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Stack>
  );
}
