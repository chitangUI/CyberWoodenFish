using System;
using System.Collections.Generic;
using System.Linq;
using UnityEngine;
using UnityEngine.UI;
using DG.Tweening;
using UnityEngine.SceneManagement;
using GooglePlayGames;
using GooglePlayGames.BasicApi;

namespace ElectronicFish
{
	public class MainManager : MonoBehaviour
	{
		[SerializeField] private Button fishImage;
		[SerializeField] private AudioSource audioSource;
		[SerializeField] private Text comboText;
		private GameObject _meritText;
		private float _lastPressTime;
		private int _combo;
		private bool _saveCombo;

		private readonly List<int> _historyCombo = new();

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

		private void Update()
		{
			if (Input.GetKeyDown(KeyCode.Escape))
			{
				SceneManager.LoadScene("WelcomeScene");
			}

			comboText.text = $"Combo: {_combo}";

			if (_historyCombo.Count == 6) // 最大为6
			{
				_historyCombo.Clear();
				return;
			}

			var time = Time.time;

			// 如果连续点击时间间隔大于 1s, 则将目前的 combo 保存到历史记录中
			if (time - _lastPressTime > 1f)
			{
				if (_combo > 0)
				{
					_historyCombo.Add(_combo);
					Social.ReportScore(_combo, GPGSIds.leaderboard, (bool success) => {
						// handle success or failure
					});
				}

				_combo = 0;
			}


			if (_historyCombo.Count == 6)
			{
				Debug.Log("You win!");
			}
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

			_lastPressTime = Time.time;

			audioSource.Play();

			_combo++;
		}
	}
}
