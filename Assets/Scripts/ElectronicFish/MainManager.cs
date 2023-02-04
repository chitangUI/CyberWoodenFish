using UnityEngine;
using UnityEngine.UI;
using DG.Tweening;

namespace ElectronicFish
{
	public class MainManager : MonoBehaviour
	{
		[SerializeField] private Button fishImage;
		private GameObject _meritText;

		// Start is called before the first frame update
		private void Awake()
		{
			fishImage.onClick.AddListener(MeritAdd);

			// 获取木鱼右上角位置的 Vector3
			var textTransformPosition = fishImage.transform.position + new Vector3(300, 500, 0);


			// 动态新建文本，字号 200, 内容为 功德+1, 颜色为白色, 上下左右居中对齐，位置为 209,443,0, 字体为Arial
			// 爱来自 Github Copilot
			_meritText = new GameObject("MeritText");
			var text = _meritText.AddComponent<Text>();
			text.fontSize = 120;
			text.text = "功德+1";
			text.color = Color.white;
			text.transform.position = textTransformPosition;
			text.alignment = TextAnchor.MiddleCenter;
			text.font = Resources.GetBuiltinResource<Font>("Arial.ttf");
			// 配置 context size filter
			var contextSizeFilter = _meritText.AddComponent<ContentSizeFitter>();
			contextSizeFilter.horizontalFit = ContentSizeFitter.FitMode.PreferredSize;
			contextSizeFilter.verticalFit = ContentSizeFitter.FitMode.PreferredSize;
			text.enabled = false;
		}

		private void MeritAdd()
		{
			var canvas = GetComponent<Canvas>();
			var newText = Instantiate(_meritText, fishImage.transform.position + new Vector3(300, 700, 0),  new Quaternion());
			newText.transform.SetParent(canvas.transform);
			newText.GetComponent<Text>().enabled = true;

			newText.transform.DOMoveY(1500, 1f).OnComplete(() =>
			{
				Destroy(newText);
			});

			PlayerPrefs.SetInt("Merit", PlayerPrefs.GetInt("Merit") + 1);
		}
	}
}
